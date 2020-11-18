package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"example.org/services/login/internal/users"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	mimeApplicationJSON = "application/json"
)

var (
	userDb = users.MakeUserDB()

	jwtSecret   = "secret"
	jwtValidity = time.Minute * 15
)

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	jsonStr, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusUnprocessableEntity)
		return
	}

	var user users.User
	err = json.Unmarshal(jsonStr, &user)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusUnprocessableEntity)
		return
	}

	err = userDb.CreateUser(user.Name, user.Pass)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	type Login struct {
		Name string `json:"user"`
		Pass string `json:"password"`
	}

	var creds Login
	err := readJSON(r.Body, &creds)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusUnprocessableEntity)
		return
	}

	user, err := userDb.GetUser(creds.Name)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	if user.Pass != creds.Pass {
		http.Error(w, "Bad user or password", http.StatusUnauthorized)
		return
	}

	token, err := createLoginToken(user)
	if err != nil {
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	type Token struct {
		Token string `json:"token"`
	}

	tokenMsg := &Token{Token: token}
	err = writeJSON(w, tokenMsg)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", mimeApplicationJSON)
}

func createLoginToken(user *users.User) (string, error) {
	claims := jwt.MapClaims{}
	claims["user"] = user.Name
	claims["exp"] = time.Now().Add(jwtValidity).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(jwtSecret))

	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func readJSON(in io.Reader, out interface{}) error {
	jsonStr, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonStr, out)
	if err != nil {
		return err
	}
	return nil
}

func writeJSON(out io.Writer, in interface{}) error {
	json, err := json.Marshal(in)
	if err != nil {
		return err
	}
	out.Write(json)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	var port int

	flag.IntVar(&port, "port", 8080, "The port to listen on")
	flag.Parse()

	router := mux.NewRouter().StrictSlash(true)

	var createUser http.Handler = http.HandlerFunc(handleCreateUser)
	createUser = handlers.MethodHandler{http.MethodPost: createUser}
	createUser = handlers.ContentTypeHandler(createUser, mimeApplicationJSON)
	createUser = handlers.LoggingHandler(os.Stdout, createUser)
	router.Handle("/users", createUser)

	var login http.Handler = http.HandlerFunc(handleLogin)
	login = handlers.MethodHandler{http.MethodPost: login}
	login = handlers.ContentTypeHandler(login, mimeApplicationJSON)
	login = handlers.LoggingHandler(os.Stdout, login)
	router.Handle("/login", login)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), router))
}

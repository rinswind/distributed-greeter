package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"example.org/services/login/internal/users"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	mimeApplicationJSON = "application/json"
)

// LoginEndpoint is the REST endpoint for the login service
type LoginEndpoint struct {
	router *mux.Router

	users *users.UserDB

	jwtSecret   string
	jwtValidity time.Duration
}

// LoginEndpoint is an http.Handler
func (lep *LoginEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lep.router.ServeHTTP(w, r)
}

// MakeLoginEndpoint creates a REST endpoint for the login service
func MakeLoginEndpoint(jwts string, jwtv time.Duration, db *users.UserDB) *LoginEndpoint {
	lep := &LoginEndpoint{users: db, jwtSecret: jwts, jwtValidity: jwtv}

	router := mux.NewRouter().StrictSlash(true)

	var createUser http.Handler = http.HandlerFunc(lep.handleCreateUser)
	createUser = handlers.MethodHandler{http.MethodPost: createUser}
	createUser = handlers.ContentTypeHandler(createUser, mimeApplicationJSON)
	router.Handle("/users", createUser)

	var login http.Handler = http.HandlerFunc(lep.handleLogin)
	login = handlers.MethodHandler{http.MethodPost: login}
	login = handlers.ContentTypeHandler(login, mimeApplicationJSON)
	router.Handle("/login", login)

	lep.router = router
	return lep
}

func (lep *LoginEndpoint) handleCreateUser(w http.ResponseWriter, r *http.Request) {
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

	err = lep.users.CreateUser(user.Name, user.Pass)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
}

func (lep *LoginEndpoint) handleLogin(w http.ResponseWriter, r *http.Request) {
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

	user, err := lep.users.GetUser(creds.Name)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	if user.Pass != creds.Pass {
		http.Error(w, "Bad user or password", http.StatusUnauthorized)
		return
	}

	token, err := lep.createLoginToken(user)
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

func (lep *LoginEndpoint) createLoginToken(user *users.User) (string, error) {
	claims := jwt.MapClaims{}
	claims["user"] = user.Name
	claims["exp"] = time.Now().Add(lep.jwtValidity).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(lep.jwtSecret))

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

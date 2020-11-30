package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"example.org/services/login/internal/auth"
	"example.org/services/login/internal/users"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	mimeApplicationJSON = "application/json"
)

var (
	router = mux.NewRouter().StrictSlash(true)
)

func init() {
	var createUser http.Handler = http.HandlerFunc(handleCreateUser)
	createUser = handlers.MethodHandler{http.MethodPost: createUser}
	createUser = handlers.ContentTypeHandler(createUser, mimeApplicationJSON)
	router.Handle("/users", createUser)

	var login http.Handler = http.HandlerFunc(handleLogin)
	login = handlers.MethodHandler{http.MethodPost: login}
	login = handlers.ContentTypeHandler(login, mimeApplicationJSON)
	router.Handle("/login", login)

	var logout http.Handler = http.HandlerFunc(handleLogout)
	logout = handlers.MethodHandler{http.MethodPost: logout}
	logout = handlers.ContentTypeHandler(logout, mimeApplicationJSON)
	router.Handle("/logout", logout)
}

// ServeHTTP is the root of the Login REST endpoint
// Implementation of http.Handler
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	jsonStr, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusUnprocessableEntity)
		return
	}

	type CreateUserRequest struct {
		UserName string `json:"user_name"`
		UserPass string `json:"user_password"`
	}

	var userMsg CreateUserRequest
	err = json.Unmarshal(jsonStr, &userMsg)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusUnprocessableEntity)
		return
	}

	userid, err := users.CreateUser(userMsg.UserName, userMsg.UserPass)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	type UserInfo struct {
		UserID   uint64 `json:"user_id"`
		UserName string `json:"user_name"`
	}

	userInfoMsg := &UserInfo{UserID: userid, UserName: userMsg.UserName}
	err = writeJSON(w, userInfoMsg)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", mimeApplicationJSON)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		UserName string `json:"user_name"`
		UserPass string `json:"user_password"`
	}

	var loginMsg LoginRequest
	err := readJSON(r.Body, &loginMsg)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusUnprocessableEntity)
		return
	}

	user, err := users.GetUserByName(loginMsg.UserName)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	if user.Pass != loginMsg.UserPass {
		http.Error(w, "Bad user or password", http.StatusUnauthorized)
		return
	}

	token, err := auth.CreateToken(user.ID, user.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = auth.CreateAuth(user.ID, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	type LoginInfo struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	tokenMsg := &LoginInfo{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken}
	err = writeJSON(w, tokenMsg)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", mimeApplicationJSON)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	type TokenInfo struct {
		AccessToken string `json:"access_token"`
	}

	var tokenInfo TokenInfo
	err := readJSON(r.Body, &tokenInfo)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusUnprocessableEntity)
		return
	}

	_, err = auth.DeleteAuth(tokenInfo.AccessToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
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

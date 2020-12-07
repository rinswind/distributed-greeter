package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

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
	router.Handle("/users", handlers.MethodHandler{
		http.MethodPost: handlers.ContentTypeHandler(
			http.HandlerFunc(handleCreateUser), mimeApplicationJSON),
		http.MethodGet: http.HandlerFunc(handleListUsers)})

	router.Handle("/users/{id}", handlers.MethodHandler{
		http.MethodGet:    http.HandlerFunc(handleUserInfo),
		http.MethodDelete: http.HandlerFunc(handleUserDelete)})

	router.Handle("/login", handlers.MethodHandler{
		http.MethodPost: handlers.ContentTypeHandler(
			http.HandlerFunc(handleLogin), mimeApplicationJSON)})

	router.Handle("/logout", handlers.MethodHandler{
		http.MethodPost: handlers.ContentTypeHandler(
			http.HandlerFunc(handleLogout), mimeApplicationJSON)})
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

	type UserCreds struct {
		Name     string `json:"user_name"`
		Password string `json:"user_password"`
	}

	var userMsg UserCreds
	err = json.Unmarshal(jsonStr, &userMsg)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusUnprocessableEntity)
		return
	}

	userid, err := users.CreateUser(userMsg.Name, userMsg.Password)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	type UserInfo struct {
		ID   uint64 `json:"user_id"`
		Name string `json:"user_name"`
	}

	userInfoMsg := &UserInfo{ID: userid, Name: userMsg.Name}
	err = writeJSON(w, userInfoMsg)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", mimeApplicationJSON)
}

func handleListUsers(w http.ResponseWriter, r *http.Request) {
	type Users struct {
		IDs []uint64 `json:"user_ids"`
	}

	userIdsMsg := &Users{IDs: *users.ListUserIDs()}

	err := writeJSON(w, userIdsMsg)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", mimeApplicationJSON)
}

func handleUserInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	user, err := users.GetUserByID(id)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusNotFound)
		return
	}

	type UserInfo struct {
		ID   uint64 `json:"user_id"`
		Name string `json:"user_name"`
	}

	userInfoMsg := &UserInfo{ID: user.ID, Name: user.Name}
	err = writeJSON(w, userInfoMsg)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", mimeApplicationJSON)
}

func handleUserDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	err = users.DeleteUser(id)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusNotFound)
		return
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	type UserCreds struct {
		Name     string `json:"user_name"`
		Password string `json:"user_password"`
	}

	var loginMsg UserCreds
	err := readJSON(r.Body, &loginMsg)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusUnprocessableEntity)
		return
	}

	user, err := users.GetUserByName(loginMsg.Name)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	if user.Pass != loginMsg.Password {
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

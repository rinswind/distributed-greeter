package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"

	"example.org/services/greeter/internal/auth"
	"example.org/services/greeter/internal/messages"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	mimeApplicationJSON = "application/json"
)

var (
	authzSplit = regexp.MustCompile("Bearer (.+)")

	router = mux.NewRouter().StrictSlash(true)
)

// MakeGreeterEndpoint creates a router for the greeter REST endpoint
func init() {
	router.Handle("/greetings",
		auth.Handler(
			handlers.MethodHandler{
				http.MethodGet: http.HandlerFunc(handleAllGreeters)}))

	router.Handle("/greetings/{lang}",
		auth.Handler(
			handlers.MethodHandler{
				http.MethodGet: http.HandlerFunc(handleGreeter)}))
}

// ServeHTTP makes the greetings endpoint an http.Handler
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}

func handleAllGreeters(w http.ResponseWriter, r *http.Request) {
	type Languages struct {
		Langs map[string]string `json:"languages"`
	}

	langs := Languages{Langs: make(map[string]string)}
	for ln := range messages.Greeters {
		langs.Langs[ln] = r.RequestURI + "/" + ln
	}

	err := writeJSON(w, langs)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", mimeApplicationJSON)
}

func handleGreeter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]

	authClaims := r.Context().Value(auth.Ctx{}).(map[string]interface{})
	user, _ := authClaims["user_name"].(string)

	msg := messages.Greeters[lang](user)

	type Message struct {
		Language string `json:"language"`
		Message  string `json:"message"`
	}

	message := Message{Language: lang, Message: msg}

	err := writeJSON(w, message)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", mimeApplicationJSON)
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

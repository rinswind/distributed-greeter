package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"

	"example.org/services/greeter/internal/messages"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	mimeApplicationJSON = "application/json"
)

var (
	authzSplit = regexp.MustCompile("Bearer (.+)")
)

// GreeterEndpoint is the REST endpoint for the greeter service
type GreeterEndpoint struct {
	messages map[string]messages.Greeter
	router   *mux.Router
}

// userCtx is a typesafe key under which the user name is attached to the request context
type userCtx struct {
}

// GreeterEndpoint is an http.Handler
func (g *GreeterEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.router.ServeHTTP(w, r)
}

// MakeGreeterEndpoint creates a router for the greeter REST endpoint
func MakeGreeterEndpoint(jwts jwt.Keyfunc, msg map[string]messages.Greeter) *GreeterEndpoint {
	ep := &GreeterEndpoint{messages: msg}

	router := mux.NewRouter().StrictSlash(true)

	var allGreeters http.Handler = http.HandlerFunc(ep.handleAllGreeters)
	allGreeters = handlers.MethodHandler{http.MethodGet: allGreeters}
	allGreeters = authHandler(jwts, allGreeters)
	router.Handle("/greetings", allGreeters)

	var greeter http.Handler = http.HandlerFunc(ep.handleGreeter)
	greeter = handlers.MethodHandler{http.MethodGet: greeter}
	greeter = authHandler(jwts, greeter)
	router.Handle("/greetings/{lang}", greeter)

	ep.router = router
	return ep
}

func (g *GreeterEndpoint) handleAllGreeters(w http.ResponseWriter, r *http.Request) {
	type Languages struct {
		Langs map[string]string `json:"languages"`
	}

	langs := Languages{Langs: make(map[string]string)}
	for ln := range g.messages {
		langs.Langs[ln] = r.RequestURI + "/" + ln
	}

	err := writeJSON(w, langs)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", mimeApplicationJSON)
}

func (g *GreeterEndpoint) handleGreeter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]
	user, _ := r.Context().Value(userCtx{}).(string)
	msg := g.messages[lang](user)

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

func authHandler(key jwt.Keyfunc, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authzHeader := r.Header.Get("Authorization")

		split := authzSplit.FindStringSubmatch(authzHeader)
		if len(split) != 2 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenStr, err := base64.StdEncoding.DecodeString(split[1])
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(string(tokenStr), key)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, ok := claims["user"]
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userCtx{}, user)
		req := r.Clone(ctx)

		next.ServeHTTP(w, req)
	})
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

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type greeter func(string) string

type userCtx struct {
}

const (
	mimeApplicationJSON = "application/json"
)

var (
	greeters map[string]greeter

	authzSplit = regexp.MustCompile("Bearer (.+)")
	jwtSecret  = "secret"
)

func init() {
	greeters = make(map[string]greeter)
	greeters["en"] = func(who string) string { return "Hello " + who }
	greeters["fr"] = func(who string) string { return "Bonjour " + who }
	greeters["bg"] = func(who string) string { return "Здравей " + who }
}

func handleAllGreeters(w http.ResponseWriter, r *http.Request) {
	type Languages struct {
		Langs map[string]string `json:"languages"`
	}

	langs := Languages{Langs: make(map[string]string)}
	for ln := range greeters {
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
	user, _ := r.Context().Value(userCtx{}).(string)
	msg := greeters[lang](user)

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

func main() {
	var port int

	flag.IntVar(&port, "port", 8080, "The port to listen on")
	flag.Parse()

	jwtKey := func(t *jwt.Token) (interface{}, error) { return []byte(jwtSecret), nil }

	router := mux.NewRouter().StrictSlash(true)

	var allGreeters http.Handler = http.HandlerFunc(handleAllGreeters)
	allGreeters = handlers.MethodHandler{http.MethodGet: allGreeters}
	allGreeters = authHandler(jwtKey, allGreeters)
	router.Handle("/greetings", allGreeters)

	var greeter http.Handler = http.HandlerFunc(handleGreeter)
	greeter = handlers.MethodHandler{http.MethodGet: greeter}
	greeter = authHandler(jwtKey, greeter)
	router.Handle("/greetings/{lang}", greeter)

	iface := fmt.Sprintf(":%v", port)
	log.Println("Starting to listen on ", iface)
	log.Fatal(http.ListenAndServe(iface, handlers.LoggingHandler(os.Stdout, router)))
}

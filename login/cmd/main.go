package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"example.org/services/login/internal/server"

	"example.org/services/login/internal/users"
	"github.com/gorilla/handlers"
)

const (
	mimeApplicationJSON = "application/json"
)

var (
	userDb = users.MakeUserDB()

	jwtSecret   = "secret"
	jwtValidity = time.Minute * 15
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "The port to listen on")
	flag.Parse()

	jwtSecret := os.Getenv("JWT_SECRET")
	jwtValidity := time.Minute * 15

	lep := server.MakeLoginEndpoint(jwtSecret, jwtValidity, users.MakeUserDB())

	iface := fmt.Sprintf(":%v", port)
	log.Println("Starting to listen on ", iface)
	log.Fatal(http.ListenAndServe(iface, handlers.LoggingHandler(os.Stdout, lep)))
}

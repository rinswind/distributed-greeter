package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"example.org/services/greeter/internal/messages"
	"example.org/services/greeter/internal/server"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "The port to listen on")
	flag.Parse()

	jwtSecret := os.Getenv("JWT_SECRET")

	jwtKey := func(t *jwt.Token) (interface{}, error) { return []byte(jwtSecret), nil }

	ep := server.MakeGreeterEndpoint(jwtKey, messages.Greeters)

	iface := fmt.Sprintf(":%v", port)
	log.Println("Starting to listen on ", iface)
	log.Fatal(http.ListenAndServe(iface, handlers.LoggingHandler(os.Stdout, ep)))
}

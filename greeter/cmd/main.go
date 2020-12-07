package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"example.org/services/greeter/internal/server"

	"github.com/gorilla/handlers"
)

func main() {
	port := os.Getenv("HTTP_PORT")
	iface := fmt.Sprintf(":%v", port)
	log.Println("Starting to listen on ", iface)
	log.Fatal(http.ListenAndServe(iface,
		handlers.LoggingHandler(os.Stdout, http.HandlerFunc(server.ServeHTTP))))
}

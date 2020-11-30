package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"example.org/services/greeter/internal/server"

	"github.com/gorilla/handlers"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "The port to listen on")
	flag.Parse()

	iface := fmt.Sprintf(":%v", port)
	log.Println("Starting to listen on ", iface)
	log.Fatal(http.ListenAndServe(iface, handlers.LoggingHandler(os.Stdout, http.HandlerFunc(server.ServeHTTP))))
}

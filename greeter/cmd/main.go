package main

import (
	"fmt"
	"os"

	"example.org/services/greeter/internal/server"
)

func main() {
	port := os.Getenv("HTTP_PORT")
	iface := fmt.Sprintf(":%v", port)
	server.Run(iface)
}

package main

import (
	"github.com/binsabit/jetinno-kapsi/internal/services"
	"log"
)

func main() {
	server, err := services.NewServer("4000")
	if err != nil {
		log.Fatal(err)
	}

	go server.RunTCPServer()

	log.Fatal(server.RunHTTPServer("3000"))
}

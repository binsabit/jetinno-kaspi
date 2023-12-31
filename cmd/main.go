package main

import (
	"github.com/binsabit/jetinno-kapsi/config"
	"github.com/binsabit/jetinno-kapsi/internal/services"
	"log"
)

func main() {

	_, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	server, err := services.NewServer("4000")
	if err != nil {
		log.Fatal(err)
	}
	go server.TCPServer.RunTCPServer()

	log.Fatal(server.RunHTTPServer("3000"))
}

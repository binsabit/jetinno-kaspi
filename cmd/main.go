package main

import (
	"errors"
	"github.com/binsabit/jetinno-kapsi/config"
	"github.com/binsabit/jetinno-kapsi/internal/services"
	"log"
	"os"
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
	if _, err := os.Stat("/logs"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("/logs", os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	go server.RunTCPServer()

	log.Fatal(server.RunHTTPServer("3000"))
}

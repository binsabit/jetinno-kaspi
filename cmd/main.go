package main

import (
	"context"
	"github.com/binsabit/jetinno-kapsi/config"
	"github.com/binsabit/jetinno-kapsi/internal/db"
	"github.com/binsabit/jetinno-kapsi/internal/services"
	"log"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	services.KASPI_QR_URL = cfg.KASPI_QR_URL
	err = db.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	tcpServer, err := services.NewTCPServer(cfg.TCPPort)

	server, err := services.NewServer(tcpServer)
	if err != nil {
		log.Fatal(err)
	}
	go server.TCPServer.RunTCPServer()

	log.Fatal(server.RunHTTPServer("3000"))
}

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
	services.KaspiLogin = cfg.KASPI_LOGIN
	services.KaspiPassword = cfg.KASPI_PASSWORD
	services.KaspiRefundURL = cfg.KASPI_REFUND_URL
	services.KaspiRefundDuration = cfg.REFUND_TIME
	err = db.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	tcpServer, err := services.NewTCPServer(cfg.TCPPort)
	if err != nil {
		log.Fatal(err)
	}
	server, err := services.NewServer(tcpServer)
	if err != nil {
		log.Fatal(err)
	}
	go server.TCPServer.RunTCPServer()

	log.Fatal(server.RunHTTPServer(cfg.HTTPPort))
}

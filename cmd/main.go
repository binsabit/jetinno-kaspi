package main

import (
	"github.com/binsabit/jetinno-kapsi/internal/services"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"log"
)

func main() {
	//go tcp server
	tcpServer, err := services.NewServer("4000")
	if err != nil {
		log.Fatal(err)
	}

	go tcpServer.AcceptConnections()
	//run http server
	log.Fatal(startHTTPServer())
}

func startHTTPServer() error {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
		JSONDecoder:           sonic.Unmarshal,
		JSONEncoder:           sonic.Marshal,
	})

	app.Use(helmet.New(helmet.ConfigDefault))
	app.Use(recover.New())

	app.Get("pay", services.WebHookHandler)
	app.Get("health", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})
	return app.Listen(":3000")
}

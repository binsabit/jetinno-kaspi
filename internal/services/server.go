package services

import (
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Server struct {
	TCPServer *TCPServer
	*fiber.App
}

type Message struct {
	Client  *Client
	Request JetinnoPayload
}

func NewServer(tcpServer *TCPServer) (*Server, error) {

	httpListener := fiber.New(fiber.Config{
		JSONDecoder: sonic.Unmarshal,
		JSONEncoder: sonic.Marshal,
	})

	httpListener.Use(helmet.New(helmet.ConfigDefault))
	httpListener.Use(recover.New())

	return &Server{
		TCPServer: tcpServer,
		App:       httpListener,
	}, nil
}

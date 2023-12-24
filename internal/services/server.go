package services

import (
	"context"
	"fmt"
	"github.com/binsabit/jetinno-kapsi/internal/db"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"net"
	"sync/atomic"
)

var clientCount atomic.Int32

type Server struct {
	TCPServer  *net.TCPListener
	TCPClients map[int64]*Client
	HTTPServer *fiber.App
	connChan   chan *net.TCPConn
	doneChan   chan struct{}
	Database   *db.Database
}

func NewServer(TCPPort string) (*Server, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%s", TCPPort))
	if err != nil {
		return nil, err
	}
	tcpListener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	httpListener := fiber.New(fiber.Config{
		JSONDecoder: sonic.Unmarshal,
		JSONEncoder: sonic.Marshal,
	})

	httpListener.Use(helmet.New(helmet.ConfigDefault))
	httpListener.Use(recover.New())

	database, err := db.New(context.Background())
	if err != nil {
		return nil, err
	}

	return &Server{
		TCPServer:  tcpListener,
		TCPClients: make(map[int64]*Client),
		HTTPServer: httpListener,
		Database:   database,
		connChan:   make(chan *net.TCPConn, 100),
	}, nil
}

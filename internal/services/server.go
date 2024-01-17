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
	"sync"
	"sync/atomic"
)

var clientCount atomic.Int32

type Server struct {
	TCPServer  *TCPServer
	HTTPServer *fiber.App
	Database   *db.Database
}

type TCPServer struct {
	Listener *net.TCPListener
	Clients  *sync.Map
	ConnChan chan *net.TCPConn
	MsgChan  chan *Message
	DoneChan chan struct{}
	sync.Mutex
}

type Message struct {
	Client  *Client
	Request JetinnoPayload
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
		TCPServer: &TCPServer{
			Listener: tcpListener,
			Clients:  &sync.Map{},
			ConnChan: make(chan *net.TCPConn),
			MsgChan:  make(chan *Message),
			DoneChan: make(chan struct{}),
			Mutex:    sync.Mutex{},
		},
		HTTPServer: httpListener,
		Database:   database,
	}, nil
}

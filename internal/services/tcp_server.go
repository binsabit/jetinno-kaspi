package services

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"
	"time"
)

var clientCount atomic.Int32

type Server struct {
	listener   *net.TCPListener
	clients    map[int32]*Client
	connection chan *Client
}

type Client struct {
	VccNo     int32
	Conn      *net.TCPConn
	WriteChan chan string
}

func (c *Client) Listen() {
	for {
		packetLength := make([]byte, 5)
		_, err := c.Conn.Read(packetLength)
		if err != nil && !errors.Is(err, io.EOF) {
			log.Printf("Error while reading :%v\n", err)
		}
		time.Sleep(time.Second * 5)
		log.Printf("%d %s", c.VccNo, string(packetLength))
		c.WriteChan <- string(packetLength)
		c.WriteChan <- "defaut"
	}
}

func (c *Client) Write() {
	for {
		select {
		case msg := <-c.WriteChan:
			err := c.HandleCommand(msg, nil)
			log.Println(err)
		default:

		}
	}
}

func (c *Client) HandleCommand(cmd string, payload interface{}) error {
	return nil
}

func (c *Client) WriteToConn(msg []byte) error {
	_, err := c.Conn.Write(msg)
	log.Println(err)
	return err
}

func NewServer(port string) (*Server, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Server{
		listener:   listener,
		clients:    make(map[int32]*Client),
		connection: make(chan *Client),
	}, nil
}

func (s *Server) AcceptConnections() {
	for {
		select {
		default:
			conn, err := s.listener.AcceptTCP()
			if err != nil {
				log.Println(err)
			}
			clientCount.Add(1)
			newClient := &Client{
				VccNo:     clientCount.Load(),
				Conn:      conn,
				WriteChan: make(chan string, 10),
			}
			err = conn.SetKeepAlive(true)
			if err != nil {
				log.Println(err)
			}
			s.clients[newClient.VccNo] = newClient
			go newClient.Listen()
			go newClient.Write()
		}
	}
}

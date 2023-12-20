package services

import (
	"log"
	"net"
	"os"
)

type Client struct {
	VccNo     int32
	Conn      *net.TCPConn
	writeChan chan []byte
}

func (c *Client) Listen() {
	for {
		content := make([]byte, 2024)
		n, err := c.Conn.Read(content)
		if err != nil {
			log.Println(err)
			continue
		}

		c.WriteToConn(content[:n])
	}
}

func (c *Client) Write() {
	for {
		select {
		case content := <-c.writeChan:
			log.Println(content)
			if content == nil {
				break
			}
			file, err := os.OpenFile("content.txt", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
			if err != nil {
				log.Printf("Error while opening file %v\n", err)
			}
			bytesWritten, err := file.Write(content)
			if err != nil {
				log.Printf("Error while writing to file %v\n", err)
			}
			if bytesWritten == 0 {
				log.Printf("No content in connection")
			}

		default:

		}
	}
}

func (c *Client) HandleCommand(cmd string, payload interface{}) error {
	switch cmd {
	default:
	}
	return nil
}

func (c *Client) WriteToConn(msg []byte) {
	c.writeChan <- msg
}

func (s *Server) RunTCPServer() {
	for {

		conn, err := s.TCPServer.AcceptTCP()
		if err != nil {
			log.Println(err)
		}
		clientCount.Add(1)
		newClient := &Client{
			VccNo:     clientCount.Load(),
			Conn:      conn,
			writeChan: make(chan []byte, 10),
		}
		err = conn.SetKeepAlive(true)
		if err != nil {
			log.Println(err)
		}
		s.TCPClients[newClient.VccNo] = newClient
		go newClient.Listen()
		go newClient.Write()

	}
}

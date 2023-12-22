package services

import (
	"encoding/binary"
	"fmt"
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
		content, err := ReadFromConn(c.Conn)
		if err != nil {
			log.Println("Error while reading")
			continue
		}

		c.WriteToConn(content)
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

			file, err := os.OpenFile(fmt.Sprintf("./logs/%d.txt", clientCount.Load()), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
			if err != nil {
				log.Printf("Error while opening file %v\n", err)
				continue
			}
			bytesWritten, err := file.Write(content)
			if err != nil {
				log.Printf("Error while writing to file %v\n", err)
				continue
			}
			if bytesWritten == 0 {
				log.Printf("No content in connection: ClientID:%d", c.VccNo)
				continue
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
		oldConn, ok := s.TCPClients[newClient.VccNo]
		if ok {
			err := oldConn.Conn.Close()
			if err != nil {
				log.Println("Error while closing connection")
			}
		}
		s.TCPClients[newClient.VccNo] = newClient
		go newClient.Listen()
		go newClient.Write()

	}
}

func ReadFromConn(conn *net.TCPConn) ([]byte, error) {
	buffer := make([]byte, 4)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}
	packetSize := binary.BigEndian.Uint32(buffer[:n])

	buffer = make([]byte, packetSize-4)
	n, err = conn.Read(buffer)
	if err != nil {
		return nil, err
	}
	return buffer[:n], nil
}

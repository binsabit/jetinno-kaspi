package services

import (
	"encoding/binary"
	"fmt"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"log"
	"net"
	"os"
)

type Client struct {
	VccNo     int32
	Conn      *net.TCPConn
	writeChan chan []byte
}

func (s *Server) Listen(client *Client) {
	for {
		content, err := ReadFromConn(client.Conn)
		if err != nil {
			log.Printf("Error while reading client:%d\n error:%v", client.VccNo, err)
			client.Conn.Close()
			delete(s.TCPClients, client.VccNo)
			return
		}

		client.WriteToConn(content)
	}
}

func (c *Client) Write() {
	for {
		select {
		case content := <-c.writeChan:
			log.Println(string(content))
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

type Request struct {
	Command        string            `json:"cmd"`
	VmcNo          string            `json:"vmc_no"`
	State          *string           `json:"state,omitempty"`
	Timestamp      *string           `json:"timestamp,omitempty"`
	Login_Count    *string           `json:"login_count,omitempty"`
	Sign           *string           `json:"sign,omitempty"`
	Version        *string           `json:"version,omitempty"`
	IO_Version     *string           `json:"io_version,omitempty"`
	Carrier_Code   *string           `json:"carrier_code,omitempty"`
	Date_Time      *string           `json:"date_time,omitempty"`
	Server_List    *string           `json:"server_list,omitempty"`
	Ret            *string           `json:"ret,omitempty"`
	Status         *string           `json:"status,omitempty"`
	Supply         map[string]string `json:"supply,omitempty"`
	Time           *string           `json:"time,omitempty"`
	IsLock         *bool             `json:"islock,omitempty"`
	QR_type        *string           `json:"qr_type,omitempty"`
	Pruduct_ID     *int64            `json:"product_id,omitempty"`
	Amount         *int64            `json:"Amount,omitempty"`
	Order_No       *string           `json:"order_no,omitempty"`
	QRCode         *string           `json:"qrcode,omitempty"`
	Product_Amount *string           `json:"product_amount,omitempty"`
	PayType        *string           `json:"paytype,omitempty"`
	PayDone        *bool             `json:"paydone,omitempty"`
}

func (c *Client) HandleCommand(cmd string, payload []byte) error {
	switch cmd {
	case pkg.COMMAND_HEARDBEAT:
	case pkg.COMMAND_ERROR_REQUEST:
	case pkg.COMMAND_LOGIN_REQUEST:
	case pkg.COMMAND_MACHINESTATUS_REQUEST:
	case pkg.COMMAND_QR_REQUEST:
	case pkg.COMMAND_CHECKORDER_REQUEST:
	case pkg.COMMAND_PAYDONE_REQUEST:
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
				continue
			}
		}
		s.TCPClients[newClient.VccNo] = newClient

		go s.Listen(newClient)
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

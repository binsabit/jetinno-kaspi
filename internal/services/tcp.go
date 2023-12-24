package services

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"github.com/bytedance/sonic"
	"io"
	"log"
	"net"
	"os"
)

type Client struct {
	VmcNo  int64
	Conn   *net.TCPConn
	Server *Server
	done   chan struct{}
}

type Request struct {
	Command        string            `json:"cmd"`
	VmcNo          int64             `json:"vmc_no"`
	State          *string           `json:"state,omitempty"`
	Timestamp      *string           `json:"timestamp,omitempty"`
	Login_Count    *int64            `json:"login_count,omitempty"`
	CompId         *int              `json:"comp_id,omitempty"`
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

func (s *Server) RunTCPServer() {
	for {
		conn, err := s.TCPServer.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		newClient := &Client{
			Conn:   conn,
			done:   make(chan struct{}),
			Server: s,
			VmcNo:  1,
		}
		go newClient.ListenConnection()
	}
}

func (c *Client) ListenConnection() {
	for {
		select {
		case <-c.done:
			return
		default:
			request, err := c.ReadFromConnection(c.Conn)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				log.Println(err)
			}
			c.VmcNo = request.VmcNo
			c.Server.mutex.Lock()
			if val, ok := c.Server.TCPClients[request.VmcNo]; ok {
				val.done <- struct{}{}
			}
			c.Server.TCPClients[c.VmcNo] = c
			c.Server.mutex.Unlock()
			log.Println(c.HandleRequest(request))
		}
	}
}

func (s *Client) ReadFromConnection(conn *net.TCPConn) (*Request, error) {
	buffer := make([]byte, 4)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}
	packetSize := binary.LittleEndian.Uint32(buffer[:n])

	buffer = make([]byte, packetSize-4)
	n, err = conn.Read(buffer)
	if err != nil {
		return nil, err
	}
	var req Request
	err = sonic.ConfigFastest.Unmarshal(buffer[8:n], &req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (c *Client) Write(content *Request) error {

	if content == nil {
		return nil
	}
	file, err := os.OpenFile(fmt.Sprintf("./logs/%d.txt", c.VmcNo), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("Error while opening file %v\n", err)
		return err
	}
	bytes, err := sonic.Marshal(content)
	if err != nil {
		log.Printf("Error while marshling json, %v\n", err)
		return err
	}
	log.Println(string(bytes))
	bytesWritten, err := file.Write(bytes)
	if err != nil {
		log.Printf("Error while writing to file %v\n", err)
		return err
	}
	if bytesWritten == 0 {
		log.Printf("No content in connection: ClientID:%d", c.VmcNo)
		return err
	}
	_, _ = file.Write([]byte("\n"))
	return nil
}

func (c *Client) HandleRequest(request *Request) error {
	if request == nil {
		return fmt.Errorf("request is empty")
	}
	switch request.Command {
	case pkg.COMMAND_HEARDBEAT:
	case pkg.COMMAND_ERROR_REQUEST:
	case pkg.COMMAND_LOGIN_REQUEST:
	case pkg.COMMAND_MACHINESTATUS_REQUEST:
	case pkg.COMMAND_QR_REQUEST:
	case pkg.COMMAND_CHECKORDER_REQUEST:
	case pkg.COMMAND_PAYDONE_REQUEST:
	default:
	}
	return c.Write(request)
}

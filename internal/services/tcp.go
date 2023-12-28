package services

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"github.com/bytedance/sonic"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

type Client struct {
	ID     int
	VmcNo  int64
	Conn   *net.TCPConn
	Server *TCPServer
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
	Ret            *int              `json:"ret,omitempty"`
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

func (t *TCPServer) RunTCPServer() {
	go t.AcceptConnections()
	for {
		select {
		case conn := <-t.ConnChan:

			request, err := ReadFromConnection(conn)
			if err != nil {
				log.Println(err)
				conn.Close()
				continue
			}
			if request == nil {
				conn.Close()
				continue
			}
			client := &Client{
				ID:     rand.Int(),
				VmcNo:  request.VmcNo,
				Conn:   conn,
				Server: t,
				done:   make(chan struct{}),
			}
			if val, ok := t.Clients.Load(request.VmcNo); ok {
				val.(*Client).done <- struct{}{}
			}
			t.Clients.Store(request.VmcNo, client)
			err = client.HandleRequest(*request)
			if err != nil {
				t.Clients.Delete(client.VmcNo)
				client.done <- struct{}{}
				log.Printf("handle command %s client:%d\n err:%v\n", request.Command, request.VmcNo, err)
			}
			go client.ReadContinuouslyFromConnection()
		}
	}
}

func (t *TCPServer) AcceptConnections() {
	for {
		conn, err := t.Listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		t.ConnChan <- conn
	}
}
func (c *Client) ReadContinuouslyFromConnection() {
	for {
		select {
		case <-c.done:
			c.Conn.Close()
			c.Server.Clients.Delete(c.VmcNo)
			return
		default:
			request, err := ReadFromConnection(c.Conn)
			if err != nil {
				if errors.Is(err, io.EOF) {
					continue
				}
				c.Conn.Close()
				return
			}
			if request == nil {
				continue
			}
			err = c.HandleRequest(*request)
			if err != nil {
				log.Printf("handle command %s client:%d\n err:%v", request.Command, request.VmcNo, err)
				return
			}
		}
	}
}

func ReadFromConnection(conn *net.TCPConn) (*Request, error) {
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

func (c *Client) Write(content ...Request) error {

	file, err := os.OpenFile(fmt.Sprintf("./logs/%d.txt", c.VmcNo), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("Error while opening file %v\n", err)
		return err
	}
	for _, val := range content {

		bytes, err := sonic.Marshal(val)
		if err != nil {
			log.Printf("Error while marshling json, %v\n", err)
			return err
		}

		log.Println(c.ID, string(bytes))
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
	}
	return nil
}

func (c *Client) HandleRequest(request Request) error {

	var response Request
	switch request.Command {
	case pkg.COMMAND_HEARDBEAT:
		response = c.HB(request)
	case pkg.COMMAND_ERROR_REQUEST:
	case pkg.COMMAND_LOGIN_REQUEST:
		response = c.Login(request)
	case pkg.COMMAND_MACHINESTATUS_REQUEST:
	case pkg.COMMAND_QR_REQUEST:
		response = c.QR(request)
	case pkg.COMMAND_CHECKORDER_REQUEST:
		response = c.CheckOrder(request)
	case pkg.COMMAND_PAYDONE_REQUEST:
	}
	log.Println(c.Write(request, response))
	return c.WriteToConn(response)
}

func (c *Client) HB(request Request) Request {
	return Request{
		VmcNo:   request.VmcNo,
		Command: pkg.COMMAND_HEARDBEAT,
	}
}

func (c *Client) QR(request Request) Request {
	response := Request{
		VmcNo:    request.VmcNo,
		Command:  pkg.COMMAND_QR_RESPONSE,
		Order_No: request.Order_No,
		QR_type:  request.QR_type,
	}
	qr := "53141999967389879258033215552005483843505"
	response.QRCode = &qr
	return response
}
func (c *Client) CheckOrder(request Request) Request {
	done := true
	response := Request{
		VmcNo:    request.VmcNo,
		Command:  pkg.COMMAND_CHECKORDER_RESPONSE,
		Order_No: request.Order_No,
		QR_type:  request.QR_type,
		PayType:  request.PayType,
		PayDone:  &done,
	}

	return response
}

func (c *Client) Login(request Request) Request {
	carrierCode := "TW-00418"
	dateTime := time.Now().Format(time.DateTime)
	serverlist := "185.100.67.252"
	ret := 0
	response := Request{
		VmcNo:        request.VmcNo,
		Command:      pkg.COMMAND_LOGIN_RESPONSE,
		Carrier_Code: &carrierCode,
		Date_Time:    &dateTime,
		Server_List:  &serverlist,
		Ret:          &ret,
	}
	return response
}

func (c *Client) WriteToConn(response Request) error {
	data, err := sonic.ConfigFastest.Marshal(response)
	if err != nil {
		return err
	}
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(len(data))+4)
	data = append(bs, data...)
	_, err = c.Conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}

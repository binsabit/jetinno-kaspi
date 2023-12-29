package services

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"github.com/bytedance/sonic"
	"log"
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
	for {
		conn, err := t.Listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		conn.SetKeepAlive(true)
		go t.HandleConnection(conn)
	}
}

func (t *TCPServer) HandleConnection(conn *net.TCPConn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()

		var req Request
		err := sonic.ConfigFastest.Unmarshal([]byte(text[12:]), &req)
		if err != nil {
			log.Println(err)
			continue
		}
		client := &Client{
			VmcNo:  req.VmcNo,
			Conn:   conn,
			Server: t,
		}
		log.Println(req)
		t.Clients.Store(req.VmcNo, client)
		client.HandleRequest(req)

	}
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

func (c *Client) HandleRequest(request Request) {

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
	c.WriteToConn(response)
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
	padding := "00000000"
	binary.LittleEndian.PutUint32(bs, uint32(len(data))+12)
	data = append(bs, append([]byte(padding), data...)...)
	_, err = c.Conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}

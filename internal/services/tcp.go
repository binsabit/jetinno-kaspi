package services

import (
	"bufio"
	"encoding/binary"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"github.com/bytedance/sonic"
	"log"
	"net"
	"regexp"
	"time"
)

type Client struct {
	ID      int
	VmcNo   int64
	Conn    *net.TCPConn
	Server  *TCPServer
	Writer  *bufio.Writer
	Scanner *bufio.Scanner
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
func extractJSON(s string) ([]string, error) {
	re := regexp.MustCompile(`\{([^}]*)\}`)

	matches := re.FindAllStringSubmatch(s, -1)

	var results []string
	for _, match := range matches {
		if len(match) == 2 {
			results = append(results, match[1])
		}
	}

	return results, nil
}

func (t *TCPServer) HandleConnection(conn *net.TCPConn) {
	scanner := bufio.NewScanner(conn)
	writer := bufio.NewWriter(conn)
	defer conn.Close()
	for scanner.Scan() {
		buffer := scanner.Text()
		var req Request
		text, err := extractJSON(buffer)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, val := range text {
			log.Println(val)
			err = sonic.ConfigFastest.Unmarshal([]byte("{"+val+"}"), &req)
			if err != nil {
				log.Println(err)
				continue
			}
			client := &Client{
				VmcNo:   req.VmcNo,
				Conn:    conn,
				Scanner: scanner,
				Writer:  writer,
				Server:  t,
			}
			t.Clients.Store(req.VmcNo, client)
			response := client.HandleRequest(req)

			data, err := sonic.ConfigFastest.Marshal(response)
			if err != nil {
				log.Println(err)
				continue
			}
			bs := make([]byte, 4)
			padding := "000y0000000"
			binary.LittleEndian.PutUint32(bs, uint32(len(data))+12)
			temp := append(bs, []byte(padding)...)
			data = append(temp, data...)
			_, err = writer.Write(data)
			if err != nil {
				log.Println(err)
				continue
			}

			if req.Command == pkg.COMMAND_QR_REQUEST {
				done := true
				res := Request{
					VmcNo:          req.VmcNo,
					Command:        pkg.COMMAND_PAYDONE_REQUEST,
					Order_No:       req.Order_No,
					Product_Amount: req.Product_Amount,
					PayDone:        &done,
					PayType:        req.PayType,
				}
				data, err := sonic.ConfigFastest.Marshal(res)
				if err != nil {
					log.Println(err)
					continue
				}
				bs := make([]byte, 4)
				padding := "000y0000000"
				binary.LittleEndian.PutUint32(bs, uint32(len(data))+12)
				temp := append(bs, []byte(padding)...)
				data = append(temp, data...)
				_, err = writer.Write(data)
				if err != nil {
					log.Println(err)
					continue
				}
			}

		}
	}
}

func (c *Client) HandleRequest(request Request) Request {

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
	return response
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
	qr := "https://qr.vendmarket.kz//storage/moonshine_users/3JZpkEc55UmYCBvUqR7AwxFNspxUjFzzV1t7hFt0.png"
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

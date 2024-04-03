package services

import (
	"context"
	"fmt"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"github.com/bytedance/sonic"
	"log"
	"math/rand"
	"net"
	"os"
	"regexp"
	"sync"
	"time"
)

var KASPI_QR_URL string

type Client struct {
	ID     int
	VmcNo  int64
	Conn   *net.TCPConn
	Server *TCPServer
	logger *log.Logger
}

func NewClient(conn *net.TCPConn, server *TCPServer) *Client {
	return &Client{
		ID:     rand.Int(),
		Conn:   conn,
		Server: server,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

type JetinnoPayload struct {
	Command          string            `json:"cmd"`
	VmcNo            int64             `json:"vmc_no"`
	IsOk             *bool             `json:"isok,omitempty"`
	State            *string           `json:"state,omitempty"`
	Timestamp        *string           `json:"timestamp,omitempty"`
	ErrorDescription *string           `json:"error_description,omitempty"`
	ErrorCode        *string           `json:"error_code,omitempty"`
	Login_Count      *int64            `json:"login_count,omitempty"`
	CompId           *int              `json:"comp_id,omitempty"`
	Sign             *string           `json:"sign,omitempty"`
	Version          *string           `json:"version,omitempty"`
	IO_Version       *string           `json:"io_version,omitempty"`
	Carrier_Code     *string           `json:"carrier_code,omitempty"`
	Date_Time        *string           `json:"date_time,omitempty"`
	Server_List      *string           `json:"server_list,omitempty"`
	Ret              *int              `json:"ret,omitempty"`
	Status           *string           `json:"status,omitempty"`
	Failreason       *string           `json:"failreason,omitempty"`
	Supply           map[string]string `json:"supply,omitempty"`
	Time             *string           `json:"time,omitempty"`
	IsLock           *bool             `json:"islock,omitempty"`
	QR_type          *string           `json:"qr_type,omitempty"`
	Pruduct_ID       *int64            `json:"product_id,omitempty"`
	Amount           *int64            `json:"Amount,omitempty"`
	Order_No         *string           `json:"order_no,omitempty"`
	QRCode           *string           `json:"qrcode,omitempty"`
	Product_Amount   *int              `json:"product_amount,omitempty"`
	PayType          *string           `json:"paytype,omitempty"`
	PayDone          *bool             `json:"paydone,omitempty"`
}

type TCPServer struct {
	Listener *net.TCPListener
	Clients  *sync.Map
}

func NewTCPServer(port int) (*TCPServer, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	tcpListener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &TCPServer{
		Listener: tcpListener,
		Clients:  &sync.Map{},
	}, nil

}
func (t *TCPServer) RunTCPServer() {
	for {
		conn, err := t.Listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		conn.SetKeepAlive(true)
		client := NewClient(
			conn,
			t,
		)
		go client.HandleConnection()
	}
}

func (c *Client) HandleConnection() {

	defer func() {
		c.Conn.Close()
	}()

Outer:
	for {

		payload := []byte{}
		brackets := 0
		for {
			b := make([]byte, 1)
			_, err := c.Conn.Read(b)
			if err != nil {
				c.logger.Println(err)
				break Outer
			}
			if b[0] == '{' {
				brackets++
			}
			if brackets == 0 {
				continue
			}
			payload = append(payload, b...)
			if b[0] == '}' {
				brackets--
			}
			if brackets == 0 {
				break
			}

		}
		request := c.extractJSON(string(payload))

		for _, r := range request {

			if val, ok := c.Server.Clients.Load(r.VmcNo); ok {
				if val.(*Client).ID != c.ID {
					log.Println("Vending machine exists")
					return
				}
			}

			c.VmcNo = r.VmcNo
			c.logger.SetPrefix(fmt.Sprintf("[vcm_no: %d] ", r.VmcNo))
			c.Server.Clients.Store(r.VmcNo, c)

			response := c.HandleRequest(r)

			if response != nil {
				if err := c.Write(*response); err != nil {
					c.logger.Println(err)
					continue
				}
			}
		}
	}

	c.Server.Clients.Delete(c.VmcNo)
}

func (c *Client) Write(response JetinnoPayload) error {
	payload, err := sonic.ConfigFastest.Marshal(response)
	if err != nil {
		return err
	}

	lengthPayload := len(payload) + 12
	lengthByte := []byte{48, 48, 48, 48}

	for i := 0; i < 4; i++ {
		if lengthPayload > 255-48 {
			lengthByte[i] = uint8(255)
			lengthPayload = lengthPayload - 255 + 48
		} else {
			lengthByte[i] = uint8(lengthPayload + 48)
			lengthPayload = 0
			break
		}

	}
	padding := []byte{116, 48, 48, 48, 48, 48, 48, 48}

	data := append(lengthByte, append(padding, payload...)...)
	_, err = c.Conn.Write(data)
	if err != nil {
		return err
	}
	c.logger.Println(string(data))
	return nil
}

func (c *Client) HandleRequest(request JetinnoPayload) *JetinnoPayload {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var response *JetinnoPayload
	switch request.Command {
	case pkg.COMMAND_HEARDBEAT:
		response = c.HB(request)
	case pkg.COMMAND_ERROR_REQUEST:
		response = c.Error(request)
	case pkg.COMMAND_LOGIN_REQUEST:
		response = c.Login(request)
	case pkg.COMMAND_MACHINESTATUS_REQUEST:
		response = c.MachineStatus(request)
	case pkg.COMMAND_QR_REQUEST:
		response = c.QR(ctx, request)
	case pkg.COMMAND_CHECKORDER_REQUEST:
		response = c.CheckOrder(ctx, request)
	case pkg.COMMAND_PAYDONE_REQUEST:
	case pkg.COMMAND_PRODUCTDONE_REQUEST:
		response = c.ProductDone(ctx, request)
	}
	return response
}

func (c Client) extractJSON(s string) []JetinnoPayload {
	re := regexp.MustCompile(`\{([^}]*)\}`)

	matches := re.FindAllStringSubmatch(s, -1)

	var (
		results     []string
		jsonPayload []JetinnoPayload
	)

	for _, match := range matches {
		if len(match) == 2 {
			results = append(results, match[1])
			var temp JetinnoPayload
			i := []byte("{" + match[1] + "}")
			c.logger.Println("request:", string(i))
			err := sonic.ConfigFastest.Unmarshal(i, &temp)
			if err != nil {
				continue
			}
			jsonPayload = append(jsonPayload, temp)
		}
	}

	return jsonPayload
}

package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/binsabit/jetinno-kapsi/internal/db"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5"
	"log"
	"math/rand"
	"net"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var KASPI_QR_URL string

type Client struct {
	ID     int
	VmcNo  int64
	Conn   *net.TCPConn
	Server *TCPServer
	stopCh chan struct{}
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
		client := &Client{
			ID:     rand.Int(),
			Conn:   conn,
			Server: t,
			stopCh: make(chan struct{}),
		}
		go client.HandleConnection()
	}
}

func extractJSON(s string) []JetinnoPayload {
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
			log.Println("request:", string(i))
			err := sonic.ConfigFastest.Unmarshal(i, &temp)
			if err != nil {
				log.Println("JSON ERROR:", err)
				continue
			}
			jsonPayload = append(jsonPayload, temp)
		}
	}

	return jsonPayload
}

func (c *Client) HandleConnection() {

	defer func() {
		c.Conn.Close()
	}()

	//reader := bufio.NewReader(conn)

Outer:
	for {

		select {
		case <-c.stopCh:
			log.Println("closing connection to:", c.VmcNo)
			break Outer
		default:
			payload := []byte{}
			log.Println(c.VmcNo)
			brackets := 0
			for {
				b := make([]byte, 1)
				_, err := c.Conn.Read(b)
				if err != nil {
					log.Println(err)
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
			request := extractJSON(string(payload))

			for _, r := range request {

				if val, ok := c.Server.Clients.Load(r.VmcNo); ok {
					if val.(*Client).ID != c.ID {
						log.Println("Vending machine exists")
						return
					}
				}

				c.VmcNo = r.VmcNo
				c.Server.Clients.Store(r.VmcNo, c)

				response := c.HandleRequest(r)

				if response != nil {
					if err := c.Write(*response); err != nil {
						log.Println(err)
						continue
					}
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
	log.Println(c.ID, string(data))
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

func (c *Client) HB(request JetinnoPayload) *JetinnoPayload {
	return &JetinnoPayload{
		VmcNo:   request.VmcNo,
		Command: pkg.COMMAND_HEARDBEAT,
	}
}

func (c *Client) PayDone(ctx context.Context, order db.Order) *JetinnoPayload {
	id, status, err := db.Storage.GetVmdIDByNo(ctx, order.VendingMachineNo)
	if err != nil {
		log.Println(err)
		return nil
	}
	if status != 1 {
		return nil
	}
	log.Println("order number", order.OrderNo)
	err = db.Storage.UpdateOrder(ctx, id, order.OrderNo, 1)
	if err != nil {
		log.Println(err)
		return nil
	}

	vmcNo, _ := strconv.ParseInt(order.VendingMachineNo, 10, 64)
	amount := int(order.Amount * 100)
	return &JetinnoPayload{
		VmcNo:          vmcNo,
		Command:        pkg.COMMAND_PAYDONE_REQUEST,
		Product_Amount: &amount,
		Order_No:       &order.OrderNo,
		PayDone:        &order.Paid,
		PayType:        &order.QRType,
	}
}

func (c *Client) QR(ctx context.Context, request JetinnoPayload) *JetinnoPayload {
	id, status, err := db.Storage.GetVmdIDByNo(ctx, strconv.FormatInt(request.VmcNo, 10))
	if err != nil {
		log.Println(err)
		return nil
	}

	if status != 1 {
		return nil
	}

	_, err = db.Storage.GetOrder(ctx, strconv.FormatInt(request.VmcNo, 10), *request.Order_No)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Println(err)
		return nil
	}

	if err == nil {
		return nil
	}

	orderID, err := db.Storage.CreateOrder(ctx, db.Order{
		OrderNo:          *request.Order_No,
		VendingMachineID: id,
		ProductID:        *request.Pruduct_ID,
		QRType:           *request.QR_type,
		Amount:           float32(*request.Amount),
	})

	if err != nil {
		log.Println(err)
		return nil
	}
	response := &JetinnoPayload{
		VmcNo:    request.VmcNo,
		Command:  pkg.COMMAND_QR_RESPONSE,
		Amount:   request.Amount,
		Order_No: request.Order_No,
		QR_type:  request.QR_type,
	}

	qr := fmt.Sprintf("%s=%d", KASPI_QR_URL, orderID)

	response.QRCode = &qr

	return response

}
func (c *Client) CheckOrder(ctx context.Context, request JetinnoPayload) *JetinnoPayload {

	order, err := db.Storage.GetOrder(ctx, strconv.FormatInt(request.VmcNo, 10), *request.Order_No)
	if err != nil {
		log.Println(err)
		return nil
	}

	amount := int64(order.Amount)

	if order.Paid && order.Status == 2 {
		return nil
	}

	if order.Paid && order.Status == 0 {
		id, status, err := db.Storage.GetVmdIDByNo(ctx, strconv.FormatInt(request.VmcNo, 10))
		if err != nil {
			log.Println(err)
			return nil
		}
		if status != 1 {
			return nil
		}

		err = db.Storage.UpdateOrder(ctx, id, *request.Order_No, 1)
		if err != nil {
			return nil
		}

	}
	response := &JetinnoPayload{
		VmcNo:    request.VmcNo,
		Command:  pkg.COMMAND_CHECKORDER_RESPONSE,
		Order_No: request.Order_No,
		Amount:   &amount,
		PayType:  &order.QRType,
		PayDone:  &order.Paid,
	}

	return response
}

func (c Client) Error(request JetinnoPayload) *JetinnoPayload {
	ctx := context.Background()

	err := db.Storage.UpdateMachineStatus(ctx, strconv.FormatInt(request.VmcNo, 10), 3)
	if err != nil {
		for {
			err = db.Storage.UpdateMachineStatus(ctx, strconv.FormatInt(request.VmcNo, 10), 3)
			if err == nil {
				break
			}
			log.Println(err)
		}
	}
	//save error
	id, _, err := db.Storage.GetVmdIDByNo(ctx, strconv.FormatInt(request.VmcNo, 10))
	if err != nil {
		for {
			id, _, err = db.Storage.GetVmdIDByNo(ctx, strconv.FormatInt(request.VmcNo, 10))
			if err == nil {
				break
			}
			log.Println(err)
		}
	}
	err = db.Storage.CreateError(ctx, id, *request.ErrorCode, *request.ErrorDescription)
	if err != nil {
		for {
			err = db.Storage.CreateError(ctx, id, *request.ErrorCode, *request.ErrorDescription)
			if err == nil {
				break
			}
			log.Println(err)
		}
	}

	return &JetinnoPayload{VmcNo: request.VmcNo, Command: pkg.COMMAND_ERROR_RESPONSE}

}

func (c *Client) MachineStatus(request JetinnoPayload) *JetinnoPayload {

	ctx := context.Background()
	if *request.Status == "clearerror" {
		err := db.Storage.UpdateMachineStatus(ctx, strconv.FormatInt(request.VmcNo, 10), 1)
		if err != nil {
			log.Println(err)
			return nil
		}

	}
	return nil
}

func (c *Client) ProductDone(ctx context.Context, request JetinnoPayload) *JetinnoPayload {

	id, _, err := db.Storage.GetVmdIDByNo(ctx, strconv.FormatInt(request.VmcNo, 10))
	if err != nil {
		log.Println(err)
		return nil
	}

	if *request.IsOk == true {
		err = db.Storage.UpdateOrder(ctx, id, *request.Order_No, 2)
		if err != nil {
			log.Println(err)
			return nil
		}
	}
	if !*request.IsOk && request.Failreason != nil {
		err = db.Storage.CreateError(ctx, request.VmcNo, *request.Failreason, "")
		if err != nil {
			for {
				err = db.Storage.CreateError(ctx, request.VmcNo, *request.Failreason, "")
				if err == nil {
					break
				}
			}
		}
	}
	response := &JetinnoPayload{
		VmcNo:    request.VmcNo,
		Command:  pkg.COMMAND_PRODUCTDONE_RESPONSE,
		Order_No: request.Order_No,
	}

	return response
}

func (c *Client) Login(request JetinnoPayload) *JetinnoPayload {
	carrierCode := "jn9527"
	dateTime := time.Now().Format(time.DateTime)
	serverlist := "185.100.67.252"
	ret := 0

	response := &JetinnoPayload{
		VmcNo:        request.VmcNo,
		Command:      pkg.COMMAND_LOGIN_RESPONSE,
		Carrier_Code: &carrierCode,
		Date_Time:    &dateTime,
		Server_List:  &serverlist,
		Ret:          &ret,
	}
	return response
}

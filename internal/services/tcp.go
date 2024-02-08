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
}

type JetinnoPayload struct {
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
	Product_Amount *int              `json:"product_amount,omitempty"`
	PayType        *string           `json:"paytype,omitempty"`
	PayDone        *bool             `json:"paydone,omitempty"`
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
		go t.HandleConnection(conn)
	}
}

func extractJSON(s string) ([]JetinnoPayload, error) {
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
			err := sonic.ConfigFastest.Unmarshal([]byte("{"+match[1]+"}"), &temp)
			if err != nil {
				log.Println(err)
				return nil, err
			}
		}
	}

	return jsonPayload, nil
}

func (t *TCPServer) HandleConnection(conn *net.TCPConn) {

	client := &Client{
		ID:     rand.Int(),
		Conn:   conn,
		Server: t,
	}

	defer func() {
		conn.Close()
		t.Clients.Delete(client.VmcNo)
	}()
	for {
		lengthByte := make([]byte, 4)

		n, err := conn.Read(lengthByte)
		if err != nil {
			log.Println(err)
			return
		}
		var length int
		for i, val := range lengthByte[:] {
			if val-48 > 0 {
				length += int(val-48) + i*48
			}

		}
		buf := make([]byte, 300)
		n, err = conn.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}
		if n < 8 {
			return
		}
		payload := buf[8:n]
		log.Println(length)
		log.Println(string(payload))
		var req JetinnoPayload
		err = sonic.ConfigFastest.Unmarshal(payload, &req)
		if err != nil {
			log.Println(err)
			return
		}
		client.VmcNo = req.VmcNo

		t.Clients.Store(req.VmcNo, client)

		response := client.HandleRequest(req)

		data, err := sonic.ConfigFastest.Marshal(response)
		if err != nil {
			log.Println(err)
			continue
		}

		if response != nil {
			log.Println(string(data))
			if err = client.Write(*response); err != nil {
				log.Println(err)
				return
			}
			if req.Command == pkg.COMMAND_QR_REQUEST {
				order := db.Order{OrderNo: *req.Order_No, VendingMachineNo: strconv.FormatInt(req.VmcNo, 10)}
				res := client.PayDone(context.Background(), order)
				if err = client.Write(*res); err != nil {
					log.Println(err)
					return
				}
			}

		}

	}

}

func (c *Client) Write(response JetinnoPayload) error {
	payload, err := sonic.ConfigFastest.Marshal(response)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
		return err
	}
	log.Println(string(data))
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
	case pkg.COMMAND_LOGIN_REQUEST:
		response = c.Login(request)
	case pkg.COMMAND_MACHINESTATUS_REQUEST:
		response = nil
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
	id, err := db.Storage.GetVmdIDByNo(ctx, order.VendingMachineNo)
	if err != nil {
		log.Println(err)
		return nil
	}
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
	id, err := db.Storage.GetVmdIDByNo(ctx, strconv.FormatInt(request.VmcNo, 10))
	if err != nil {
		log.Println(err)
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
		id, err := db.Storage.GetVmdIDByNo(ctx, strconv.FormatInt(request.VmcNo, 10))
		if err != nil {
			log.Println(err)
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
func (c *Client) ProductDone(ctx context.Context, request JetinnoPayload) *JetinnoPayload {
	id, err := db.Storage.GetVmdIDByNo(ctx, strconv.FormatInt(request.VmcNo, 10))
	if err != nil {
		log.Println(err)
		return nil
	}
	err = db.Storage.UpdateOrder(ctx, id, *request.Order_No, 2)
	if err != nil {
		log.Println(err)
		return nil
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

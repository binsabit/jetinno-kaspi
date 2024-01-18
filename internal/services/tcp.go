package services

import (
	"bufio"
	"encoding/base64"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"github.com/bytedance/sonic"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
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

	client := &Client{
		ID:     rand.Int(),
		Conn:   conn,
		Server: t,
	}
	defer conn.Close()
	for {
		lengthByte := make([]byte, 4)

		n, err := conn.Read(lengthByte)
		if err != nil {
			log.Println(err)
			return
		}
		var length int
		for _, val := range lengthByte[:] {
			length += int(val - 48)
		}
		log.Println(lengthByte[:4])
		buf := make([]byte, length)
		n, err = conn.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}
		if n < 8 {
			return
		}
		payload := buf[8:n]
		request := JetinnoPayload{}
		err = sonic.ConfigFastest.Unmarshal(payload, &request)
		if err != nil {
			log.Println(string(buf), lengthByte, buf)
			log.Println(err)
			return
		}

		client.VmcNo = request.VmcNo

		t.Clients.Store(request.VmcNo, client)

		response := client.HandleRequest(request)
		if response != nil {
			if err = client.Write(*response); err != nil {
				log.Println(err)
				continue
			}

			log.Println(client.ID, "request:", string(buf))
		}
	}
	//scanner := bufio.NewScanner(conn)
	//writer := bufio.NewWriter(conn)
	//for scanner.Scan() {
	//	buffer := scanner.Text()
	//	log.Println(buffer)
	//	var req JetinnoPayload
	//	text, err := extractJSON(buffer)
	//	if err != nil {
	//		log.Println(err)
	//		continue
	//	}
	//	for _, val := range text {
	//		err = sonic.ConfigFastest.Unmarshal([]byte("{"+val+"}"), &req)
	//		if err != nil {
	//			log.Println(err, val)
	//			continue
	//		}
	//		client := &Client{
	//			VmcNo:   req.VmcNo,
	//			Conn:    conn,
	//			Scanner: scanner,
	//			Writer:  writer,
	//			Server:  t,
	//		}
	//
	//		t.Clients.Store(req.VmcNo, client)
	//
	//		response := client.HandleRequest(req)
	//		if err = client.Write(response); err != nil {
	//			log.Println(err)
	//			continue
	//		}
	//
	//		log.Println(clientCode, "request:", val)
	//	}
	//}
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
	log.Println(c.ID, string(data))
	return nil
}

func (c *Client) HandleRequest(request JetinnoPayload) *JetinnoPayload {

	var response JetinnoPayload
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
	case pkg.COMMAND_PRODUCTDONE_REQUEST:
		response = c.ProductDone(request)
	}
	return &response
}

func (c *Client) HB(request JetinnoPayload) JetinnoPayload {
	return JetinnoPayload{
		VmcNo:   request.VmcNo,
		Command: pkg.COMMAND_HEARDBEAT,
	}
}

func (c *Client) QR(request JetinnoPayload) JetinnoPayload {
	response := JetinnoPayload{
		VmcNo:    request.VmcNo,
		Command:  pkg.COMMAND_QR_RESPONSE,
		Amount:   request.Amount,
		Order_No: request.Order_No,
		QR_type:  request.QR_type,
	}
	//qr := "https://qr.vendmarket.kz//storage/moonshine_users/3JZpkEc55UmYCBvUqR7AwxFNspxUjFzzV1t7hFt0.png"
	img, err := os.Open("1.png")
	if err != nil {
		log.Println(err)
		return response
	}
	data, err := io.ReadAll(img)
	if err != nil {
		log.Println(err)
		return response
	}
	sEnc := base64.StdEncoding.EncodeToString(data)
	response.QRCode = &sEnc
	return response
}
func (c *Client) CheckOrder(request JetinnoPayload) JetinnoPayload {
	done := true
	response := JetinnoPayload{
		VmcNo:    request.VmcNo,
		Command:  pkg.COMMAND_CHECKORDER_RESPONSE,
		Order_No: request.Order_No,
		Amount:   request.Amount,
		QR_type:  request.QR_type,
		PayType:  request.PayType,
		PayDone:  &done,
	}

	return response
}
func (c *Client) ProductDone(request JetinnoPayload) JetinnoPayload {
	response := JetinnoPayload{
		VmcNo:    request.VmcNo,
		Command:  pkg.COMMAND_PRODUCTDONE_RESPONSE,
		Order_No: request.Order_No,
	}

	return response
}

func (c *Client) Login(request JetinnoPayload) JetinnoPayload {
	carrierCode := "jn9527"
	dateTime := time.Now().Format(time.DateTime)
	serverlist := "185.100.67.252"
	ret := 0
	response := JetinnoPayload{
		VmcNo:        request.VmcNo,
		Command:      pkg.COMMAND_LOGIN_RESPONSE,
		Carrier_Code: &carrierCode,
		Date_Time:    &dateTime,
		Server_List:  &serverlist,
		Ret:          &ret,
	}
	return response
}

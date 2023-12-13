package internal

import (
	"encoding/xml"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"math/rand"
	"strconv"
	"time"
)

type KaspiWebHookRequest struct {
	Account string    `json:"account"`
	Command string    `json:"command"`
	Sum     float64   `json:"sum"`
	TxnID   int64     `json:"txn_id"`
	TxnDate time.Time `json:"txn_date"`
}

func NewKaspiWebHookRequest(query map[string]string) *KaspiWebHookRequest {

	sum, _ := strconv.ParseFloat(query["sum"], 64)
	txnID, _ := strconv.ParseInt(query["txn_id"], 10, 64)
	s := query["txn_date"]

	year := s[:4]
	month := s[4:6]
	day := s[6:8]
	hour := s[8:10]
	minute := s[10:12]
	second := s[12:]

	designatedTime := fmt.Sprintf("%s-%s-%sT%s:%s:%s", year, month, day, hour, minute, second)

	txnDate, _ := time.Parse(time.DateOnly, designatedTime)
	return &KaspiWebHookRequest{
		Account: query["account"],
		Command: query["command"],
		Sum:     sum,
		TxnID:   txnID,
		TxnDate: txnDate,
	}

}

type KaspiWebHookPayResponse struct {
	XMLName       xml.Name `xml:"response"`
	ProviderTxnID int64    `json:"prv_txn_id" xml:"prv_txn_id"`
	TxnID         int64    `json:"txn_id" xml:"txn_id"`
	Result        int8     `json:"result" xml:"result"`
	Sum           string   `json:"sum" xml:"sum"`
	Comment       string   `json:"comment" xml:"comment"`
}

func WebHookHandler(ctx *fiber.Ctx) error {
	query := ctx.Queries()
	request := NewKaspiWebHookRequest(query)

	response := KaspiWebHookPayResponse{
		Result: 5,
		TxnID:  request.TxnID,
	}

	response.Result = 0
	response.ProviderTxnID = rand.Int63()
	response.Sum = strconv.FormatFloat(request.Sum, 'f', 2, 64)

	return ctx.Status(fiber.StatusOK).XML(response)
}

type KaspiQrcodeRequest struct {
	TranId         string `json:"TranId"`
	OrderId        string `json:"OrderId"`
	Amount         int    `json:"Amount"`
	Service        string `json:"Service"`
	ReturnUrl      string `json:"returnUrl"`
	RefererHost    string `json:"refererHost"`
	GenerateQrCode bool   `json:"GenerateQrCode,omitempty"`
}

type KaspiQrcodeResponse struct {
	Code        int    `json:"code"`
	RedirectUrl string `json:"redirectUrl"`
	Message     string `json:"message"`
	QrCodeImage string `json:"qrCodeImage"`
}

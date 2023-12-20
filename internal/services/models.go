package services

import (
	"encoding/xml"
	"github.com/binsabit/jetinno-kapsi/config"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"strconv"
)

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

type KaspiWebHookRequest struct {
	Account string  `json:"account"`
	Command string  `json:"command"`
	Sum     float64 `json:"sum"`
	TxnID   int64   `json:"txn_id"`
	TxnDate int64   `json:"txn_date"`
}

func NewKaspiWebHookRequest(query map[string]string) KaspiWebHookRequest {

	sum, _ := strconv.ParseFloat(query["sum"], 64)
	txnID, _ := strconv.ParseInt(query["txn_id"], 10, 64)
	txnDate, _ := strconv.ParseInt(query["txn_date"], 10, 64)

	return KaspiWebHookRequest{
		Account: query["account"],
		Command: query["command"],
		Sum:     sum,
		TxnID:   txnID,
		TxnDate: txnDate,
	}

}

type KaspiWebHookResponse struct {
	XMLName       xml.Name `xml:"response"`
	ProviderTxnID int64    `json:"prv_txn_id" xml:"prv_txn_id"`
	TxnID         int64    `json:"txn_id" xml:"txn_id"`
	Result        int      `json:"result" xml:"result"`
	Sum           string   `json:"sum" xml:"sum"`
	Comment       string   `json:"comment" xml:"comment"`
}

type KaspiQuickPayRequest struct {
	TranId         string `json:"TranId"`
	OrderId        string `json:"OrderId"`
	Amount         int    `json:"Amount"`
	Service        string `json:"Service"`
	ReturnUrl      string `json:"returnUrl"`
	RefererHost    string `json:"refererHost"`
	GenerateQrCode bool   `json:"GenerateQrCode,omitempty"`
}

func NewKaspiQuickPayRequest(tranId string, orderId string, amount int) *KaspiQuickPayRequest {
	a := &KaspiQuickPayRequest{
		TranId:         tranId,
		OrderId:        orderId,
		Amount:         amount,
		Service:        config.AppConfig.KASPI_SERVICE_ID,
		ReturnUrl:      config.AppConfig.KASPI_REFERRED_HOST,
		RefererHost:    config.AppConfig.KASPI_REDIRECT_URL,
		GenerateQrCode: true,
	}
	return a
}

type KaspiQuickPayResponse struct {
	Code        int    `json:"code"`
	RedirectUrl string `json:"redirectUrl"`
	Message     string `json:"message"`
	QRCodeImage string `json:"qrCodeImage"`
}

func (k KaspiQuickPayResponse) IsSuccess() bool {
	return k.Code == pkg.KASPI_PAYMENT_SUCCESS

}

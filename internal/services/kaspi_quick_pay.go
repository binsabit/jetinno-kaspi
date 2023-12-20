package services

import (
	"github.com/binsabit/jetinno-kapsi/config"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

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
	return k.Code == pkg.KASPI_SUCCESS
}

func MakePayment(tranID, orderID string, amount int) (*KaspiQuickPayResponse, error) {
	var (
		requestData  = NewKaspiQuickPayRequest(tranID, orderID, amount)
		responseData KaspiQuickPayResponse
	)

	paymentRequest := pkg.Request{
		URL: config.AppConfig.KASPI_PAYMENT_URL,
		Header: map[string]string{
			"ContentType": "application/json",
		},
		Method: fiber.MethodPost,
		Data:   requestData,
	}

	data, err := paymentRequest.Do()
	if err != nil {

		//TODO handle error
		log.Info(err.Error())
		return nil, err

	}
	err = sonic.ConfigFastest.Unmarshal(data, &responseData)
	if err != nil {
		log.Info(err.Error())
		return nil, err
	}

	return &responseData, nil
}

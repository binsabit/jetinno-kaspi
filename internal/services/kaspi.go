package services

import (
	"context"
	"github.com/binsabit/jetinno-kapsi/config"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"time"
)

func (s *Server) ProcessPayment(ctx context.Context, request KaspiWebHookRequest) (int64, int, error) {

	prevTxn, err := GetTransactionBYTxnID(ctx, s.Database.GetDB(), request.TxnID)
	if err != nil {
		return 0, pkg.KASPI_PROVIDER_ERROR, err
	}
	if prevTxn != nil {
		return 0, prevTxn.Result, nil
	}

	tx, err := s.Database.GetDB().Begin(ctx)
	if err != nil {
		return 0, 5, err
	}
	defer tx.Rollback(ctx)
	provTxnID, err := CreateTransaction(ctx, tx, Transaction{
		TxnID:     request.TxnID,
		TxnDate:   request.TxnDate,
		Result:    pkg.KASPI_PAYMENT_SUCCESS,
		Sum:       request.Sum,
		Comment:   "OK",
		Status:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		return 0, pkg.KASPI_PROVIDER_ERROR, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, pkg.KASPI_PROVIDER_ERROR, err
	}

	return provTxnID, pkg.KASPI_PAYMENT_SUCCESS, nil

}

func MakeQROrder(tranID, orderID string, amount int) (*KaspiQuickPayResponse, error) {
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
		return nil, err
	}
	err = sonic.ConfigFastest.Unmarshal(data, &responseData)
	if err != nil {
		return nil, err
	}

	return &responseData, nil
}

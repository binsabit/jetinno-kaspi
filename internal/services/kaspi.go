package services

import (
	"context"
	"github.com/binsabit/jetinno-kapsi/config"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"time"
)

func (s *Server) ProcessPayment(ctx context.Context, request KaspiWebHookRequest) (int64, int, error) {

	prevTxn, err := GetTransactionBYTxnID(ctx, s.Database.GetDB(), request.TxnID)
	if err != nil {
		return 0, pkg.KASPI_PROVIDER_ERROR, err
	}
	if prevTxn == nil {
		return 0, pkg.KASPI_PAYMENT_NOTEXISTS, nil
	}

	if prevTxn.Status != pkg.PAYMENT_STATUS_CREATED {
		return prevTxn.ID, prevTxn.Result, nil
	}

	tx, err := s.Database.GetDB().Begin(ctx)
	if err != nil {
		return 0, pkg.KASPI_PROVIDER_ERROR, err
	}

	defer tx.Rollback(ctx)

	UpdateTransaction(ctx, s.Database.GetDB(), UpdateArgs{
		Result: &pkg.KASPI_PAYMENT_SUCCESS,
		Status: &pkg.PAYMENT_STATUS_PAID,
	})

	if err != nil {
		return 0, pkg.KASPI_PROVIDER_ERROR, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, pkg.KASPI_PROVIDER_ERROR, err
	}

	return prevTxn.ID, pkg.KASPI_PAYMENT_SUCCESS, nil

}

func (s *Server) MakeQROrder(ctx context.Context, tranID, orderID string, amount int) (*KaspiQuickPayResponse, error) {
	var (
		requestData  = NewKaspiQuickPayRequest(tranID, orderID, amount)
		responseData KaspiQuickPayResponse
	)

	//create order in transactions
	orderIDInt, err := strconv.ParseInt(requestData.OrderId, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = CreateTransaction(ctx, s.Database.GetDB(), Transaction{
		Status:    false,
		OrderID:   orderIDInt,
		Sum:       float64(requestData.Amount),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

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

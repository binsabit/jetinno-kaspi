package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/binsabit/jetinno-kapsi/internal/db"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"github.com/jackc/pgx/v5"
	"strconv"
)

func (c *Client) QR(ctx context.Context, request JetinnoPayload) *JetinnoPayload {
	id, status, err := db.Storage.GetVmdIDByNo(ctx, strconv.FormatInt(request.VmcNo, 10))
	if err != nil {
		c.logger.Println(err)
		return nil
	}

	if status != 1 {
		return nil
	}

	_, err = db.Storage.GetOrder(ctx, strconv.FormatInt(request.VmcNo, 10), *request.Order_No)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		c.logger.Println(err)
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
		c.logger.Println(err)
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

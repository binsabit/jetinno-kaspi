package services

import (
	"context"
	"github.com/binsabit/jetinno-kapsi/internal/db"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"strconv"
)

func (c *Client) CheckOrder(ctx context.Context, request JetinnoPayload) *JetinnoPayload {

	order, err := db.Storage.GetOrder(ctx, strconv.FormatInt(request.VmcNo, 10), *request.Order_No)
	if err != nil {
		c.logger.Println(err)
		return nil
	}

	amount := int64(order.Amount)

	if order.Paid && order.Status == 2 {
		return nil
	}

	response := &JetinnoPayload{
		VmcNo:    request.VmcNo,
		Command:  pkg.COMMAND_CHECKORDER_RESPONSE,
		Order_No: request.Order_No,
		Amount:   &amount,
		PayType:  &order.QRType,
		PayDone:  &order.Paid,
	}

	if order.Paid && order.Status == 0 {
		id, status, err := db.Storage.GetVmdIDByNo(ctx, strconv.FormatInt(request.VmcNo, 10))
		if err != nil {
			c.logger.Println(err)
			return nil
		}
		if status != 1 {
			*response.PayDone = false
			return response
		}

		err = db.Storage.UpdateOrder(ctx, id, *request.Order_No, 1)
		for err != nil {
			c.logger.Println(err)
			err = db.Storage.UpdateOrder(ctx, id, *request.Order_No, 1)
		}
	}

	return response
}

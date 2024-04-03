package services

import (
	"context"
	"github.com/binsabit/jetinno-kapsi/internal/db"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"strconv"
)

func (c *Client) ProductDone(ctx context.Context, request JetinnoPayload) *JetinnoPayload {

	id, _, err := db.Storage.GetVmdIDByNo(ctx, strconv.FormatInt(request.VmcNo, 10))
	if err != nil {
		c.logger.Println(err)
		return nil
	}

	if *request.IsOk == true {
		err = db.Storage.UpdateOrder(ctx, id, *request.Order_No, pkg.OrderUploaded)
		if err != nil {
			c.logger.Println(err)
			return nil
		}
	}

	if !*request.IsOk {
		order, err := db.Storage.GetOrder(ctx, strconv.FormatInt(request.VmcNo, 10), *request.Order_No)
		if err != nil {
			c.logger.Println("error: %v", err)
		}

		err = db.Storage.UpdateOrder(ctx, id, *request.Order_No, pkg.OrderDisrupted)
		if err != nil {
			c.logger.Println(err)
			return nil
		}
		c.Refund(c.VmcNo, order.ID)
	}

	response := &JetinnoPayload{
		VmcNo:    request.VmcNo,
		Command:  pkg.COMMAND_PRODUCTDONE_RESPONSE,
		Order_No: request.Order_No,
	}

	return response
}

package services

import (
	"context"
	"github.com/binsabit/jetinno-kapsi/internal/db"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"log"
	"strconv"
)

func (c *Client) PayDone(ctx context.Context, order db.Order) *JetinnoPayload {
	id, status, err := db.Storage.GetVmdIDByNo(ctx, order.VendingMachineNo)
	if err != nil {
		log.Println(err)
		return nil
	}
	if status != 1 {
		return nil
	}
	err = db.Storage.UpdateOrder(ctx, id, order.OrderNo, pkg.OrderPaid)
	if err != nil {
		c.logger.Println(err)
		return nil
	}

	vmcNo, _ := strconv.ParseInt(order.VendingMachineNo, 10, 64)
	amount := int(order.Amount * 100)
	return &JetinnoPayload{
		VmcNo:          vmcNo,
		Command:        pkg.COMMAND_PAYDONE_REQUEST,
		Product_Amount: &amount,
		Pruduct_ID:     &order.ProductID,
		Order_No:       &order.OrderNo,
		PayDone:        &order.Paid,
		PayType:        &order.QRType,
	}
}

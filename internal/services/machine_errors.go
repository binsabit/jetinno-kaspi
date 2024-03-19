package services

import (
	"context"
	"github.com/binsabit/jetinno-kapsi/internal/db"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"strconv"
)

func (c Client) Error(request JetinnoPayload) *JetinnoPayload {
	ctx := context.Background()

	err := db.Storage.UpdateMachineStatus(ctx, strconv.FormatInt(request.VmcNo, 10), 3)
	if err != nil {
		for {
			err = db.Storage.UpdateMachineStatus(ctx, strconv.FormatInt(request.VmcNo, 10), 3)
			if err == nil {
				break
			}
			c.logger.Println(err)
		}
	}
	id, _, err := db.Storage.GetVmdIDByNo(ctx, strconv.FormatInt(request.VmcNo, 10))
	if err != nil {
		for {
			id, _, err = db.Storage.GetVmdIDByNo(ctx, strconv.FormatInt(request.VmcNo, 10))
			if err == nil {
				break
			}
			c.logger.Println(err)
		}
	}
	err = db.Storage.CreateError(ctx, id, *request.ErrorCode, *request.ErrorDescription)
	if err != nil {
		for {
			err = db.Storage.CreateError(ctx, id, *request.ErrorCode, *request.ErrorDescription)
			if err == nil {
				break
			}
			c.logger.Println(err)
		}
	}

	order, err := db.Storage.GetLastNotUploadedOrder(ctx, strconv.FormatInt(request.VmcNo, 10))
	if err != nil {
		c.logger.Println("error while getting last order: %v", err)
		return &JetinnoPayload{VmcNo: request.VmcNo, Command: pkg.COMMAND_ERROR_RESPONSE}

	}
	c.Refund(request.VmcNo, order.ID)

	return &JetinnoPayload{VmcNo: request.VmcNo, Command: pkg.COMMAND_ERROR_RESPONSE}

}

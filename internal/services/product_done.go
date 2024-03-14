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
		err = db.Storage.UpdateOrder(ctx, id, *request.Order_No, 2)
		if err != nil {
			c.logger.Println(err)
			return nil
		}
	}
	if !*request.IsOk {
		if request.Failreason != nil {
			err = db.Storage.CreateError(ctx, id, *request.Failreason, "")
			if err != nil {
				c.logger.Println(err)
				for {
					err = db.Storage.CreateError(ctx, id, *request.Failreason, "")
					if err == nil {
						break
					}
				}
			}
		}
	}
	response := &JetinnoPayload{
		VmcNo:    request.VmcNo,
		Command:  pkg.COMMAND_PRODUCTDONE_RESPONSE,
		Order_No: request.Order_No,
	}

	return response
}

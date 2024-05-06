package services

import (
	"context"
	"github.com/binsabit/jetinno-kapsi/internal/db"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"strconv"
)

func (c Client) Error(request JetinnoPayload) *JetinnoPayload {
	ctx := context.Background()

	id, _, err := db.Storage.GetVmdIDByNo(ctx, strconv.FormatInt(request.VmcNo, 10))
	if err != nil {
		c.logger.Println(err)
		return nil
	}

	err = db.Storage.CreateError(ctx, id, *request.ErrorCode, *request.ErrorDescription)
	if err != nil {
		c.logger.Println(err)
		return nil
	}

	return &JetinnoPayload{VmcNo: request.VmcNo, Command: pkg.COMMAND_ERROR_RESPONSE}

}

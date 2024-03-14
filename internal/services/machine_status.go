package services

import (
	"context"
	"github.com/binsabit/jetinno-kapsi/internal/db"
	"strconv"
)

func (c *Client) MachineStatus(request JetinnoPayload) *JetinnoPayload {

	ctx := context.Background()
	if *request.Status == "clearerror" {
		err := db.Storage.UpdateMachineStatus(ctx, strconv.FormatInt(request.VmcNo, 10), 1)
		if err != nil {
			c.logger.Println(err)
			return nil
		}

	}
	return nil
}

package services

import (
	"context"
	"github.com/binsabit/jetinno-kapsi/internal/db"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"strconv"
	"time"
)

func (c *Client) Login(request JetinnoPayload) *JetinnoPayload {

	vendingMachine, err := db.Storage.GetVcmByNo(context.Background(), strconv.FormatInt(request.VmcNo, 10))
	if err != nil {
		return nil
	}
	carrierCode := vendingMachine.Password
	dateTime := time.Now().Format(time.DateTime)
	serverlist := "185.100.67.252"
	ret := 0

	response := &JetinnoPayload{
		VmcNo:        request.VmcNo,
		Command:      pkg.COMMAND_LOGIN_RESPONSE,
		Carrier_Code: &carrierCode,
		Date_Time:    &dateTime,
		Server_List:  &serverlist,
		Ret:          &ret,
	}
	return response
}

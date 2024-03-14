package services

import (
	"github.com/binsabit/jetinno-kapsi/pkg"
	"time"
)

func (c *Client) Login(request JetinnoPayload) *JetinnoPayload {
	carrierCode := "jn9527"
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

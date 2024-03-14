package services

import "github.com/binsabit/jetinno-kapsi/pkg"

func (c *Client) HB(request JetinnoPayload) *JetinnoPayload {
	return &JetinnoPayload{
		VmcNo:   request.VmcNo,
		Command: pkg.COMMAND_HEARDBEAT,
	}
}

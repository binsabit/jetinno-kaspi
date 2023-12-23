package pkg

var (
	KASPI_PAYMENT_SUCCESS   = 0
	KASPI_PAYMENT_NOTEXISTS = 3
	KASPI_PROVIDER_ERROR    = 5
)

var (
	PAYMENT_STATUS_CREATED = false
	PAYMENT_STATUS_PAID    = true
)

const (
	COMMAND_HEARDBEAT              = "hb"
	COMMAND_QR_REQUEST             = "qrcode"
	COMMAND_QR_RESPONSE            = "qrcode_"
	COMMAND_LOGIN_REQUEST          = "login"
	COMMAND_LOGIN_RESPONSE         = "login_r"
	COMMAND_MACHINESTATUS_REQUEST  = "machinestatus"
	COMMAND_MACHINESTATUS_RESPONSE = "machinestatus_r"
	COMMAND_ERROR_REQUEST          = "error"
	COMMAND_ERROR_RESPONSE         = "error_r"
	COMMAND_PAYDONE_REQUEST        = "pay_done"
	COMMAND_PAYDONE_RESPONSE       = "paydone_r"
	COMMAND_CHECKORDER_REQUEST     = "checkorder"
	COMMAND_CHECKORDER_RESPONSE    = "checkorder_r"
)

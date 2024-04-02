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
	COMMAND_QR_RESPONSE            = "qrcode_r"
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
	COMMAND_PRODUCTDONE_REQUEST    = "productdone"
	COMMAND_PRODUCTDONE_RESPONSE   = "productdone_r"
)

var ErrorCodeMap = map[string]string{
	"ERROR:5300": "Cup container lacks cups",
	"ERROR:5C05": "Flowmeter error/flow rate is abnormal",
	"ERROR:5700": "Bean container lacks coffee beans",
	"ERROR:5A00": "Waste bin is full/Grounds bin is not in place",
	"ERROR:5600": "There is cup in cupholder",
	"ERROR:7100": "Boiler filling timeout",
	"ERROR:0000": "Unknown Error",
	"ERROR:5406": "Cup container dispenses cup failure",
	"ERROR:5902": "Bucket 2 Lack Of Water",
	"ERROR:5B00": "Drip tray is not in place",
	"ERROR:7200": "Air break filling timeout",
}

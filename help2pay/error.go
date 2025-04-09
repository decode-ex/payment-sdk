package help2pay

import "errors"

var (
	ErrInvalidMerchantOrderID = errors.New("invalid merchant order ID")
	ErrInvalidAmount          = errors.New("invalid amount")
	ErrUnsupportedBank        = errors.New("unsupported bank")
	ErrInvalidCurrency        = errors.New("invalid currency")
	ErrInvalidCustomerID      = errors.New("invalid customer id")
	ErrInvalidCustomerIP      = errors.New("invalid customer ip")
)

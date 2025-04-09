package asiabank

import "errors"

var (
	ErrInvalidMerchantOrderID   = errors.New("invalid merchant order ID")
	ErrInvalidCurrency          = errors.New("invalid currency")
	ErrInvalidAmount            = errors.New("invalid amount")
	ErrInvalidCustomerIP        = errors.New("invalid customer IP")
	ErrInvalidCustomerFirstName = errors.New("invalid customer first name")
	ErrInvalidCustomerLastName  = errors.New("invalid customer last name")
	ErrInvalidCustomerPhone     = errors.New("invalid customer phone")
	ErrInvalidCustomerEmail     = errors.New("invalid customer email")
	ErrInvalidNetwork           = errors.New("invalid network")
)

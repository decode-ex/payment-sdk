package xpay

import "errors"

var (
	ErrorEmptyCallbackPayload                    = errors.New("empty callback payload")
	ErrorInvalidCallbackPayload                  = errors.New("invalid callback payload")
	ErrorCallbackPayloadMissingRequiredField     = errors.New("callback payload missing required field")
	ErrorEmptyCallbackPayloadData                = errors.New("empty callback payload data")
	ErrorInvalidCallbackPayloadData              = errors.New("invalid callback payload data")
	ErrorCallbackPayloadDataMissingRequiredField = errors.New("callback payload data missing required field")
	ErrInvalidSign                               = errors.New("invalid sign")
	ErrorUnsupportedCurrency                     = errors.New("unsupported currency")
	ErrorInvalidAmount                           = errors.New("invalid amount")
	ErrorInvalidData                             = errors.New("invalid data")
)

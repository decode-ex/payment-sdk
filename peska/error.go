package peska

import "errors"

type ErrorCode = int

const (
	// PAY request was successful
	ErrorCodeSucess ErrorCode = 200
	// Forbidden domain
	ErrorCodeForbidden ErrorCode = 40301
	// Invalid content type
	ErrorCodeInvalidContent ErrorCode = 40001
	// Missing header parameters
	ErrorCodeMissingHeader ErrorCode = 40002
	// Invalid timestamp
	ErrorCodeInvalidTimestamp ErrorCode = 40004
	// Currency not support
	ErrorCodeCurrencyNotSupport ErrorCode = 40005
	// Invalid transfer amount
	ErrorCodeInvalidTransferAmount ErrorCode = 40006
	// Authentication key failed
	ErrorCodeAuthFailed ErrorCode = 40101
	// Signature verification failed
	ErrorCodeSignatureFailed ErrorCode = 40102
	// Merchant account is not exist or not active
	ErrorCodeMerchantNotExist ErrorCode = 40210
	// User account is not exist or not active
	ErrorCodeUserNotExist ErrorCode = 40220
	// Merchant order repeat
	ErrorCodeMerchantOrderRepeat ErrorCode = 40910
	// Merchant order not exist
	ErrorCodeMerchantOrderNotExist ErrorCode = 40920
	// Value invalid
	ErrorCodeValueInvalid ErrorCode = 422
)

var (
	ErrForbidden             = errors.New("forbidden domain")
	ErrInvalidContent        = errors.New("invalid content type")
	ErrMissingHeader         = errors.New("missing header parameters")
	ErrInvalidTimestamp      = errors.New("invalid timestamp")
	ErrCurrencyNotSupport    = errors.New("currency not support")
	ErrInvalidTransferAmount = errors.New("invalid transfer amount")
	ErrAuthFailed            = errors.New("authentication key failed")
	ErrSignatureFailed       = errors.New("signature verification failed")
	ErrMerchantNotExist      = errors.New("merchant account is not exist or not active")
	ErrUserNotExist          = errors.New("user account is not exist or not active")
	ErrMerchantOrderRepeat   = errors.New("merchant order repeat")
	ErrMerchantOrderNotExist = errors.New("merchant order not exist")
	ErrValueInvalid          = errors.New("value invalid")
)

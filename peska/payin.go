package peska

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type PayInCurrency = string

const (
	PayInCurrencyUSD PayInCurrency = "USD"
	PayInCurrencyEUR PayInCurrency = "EUR"
	PayInCurrencyGBP PayInCurrency = "GBP"
	PayInCurrencyJPY PayInCurrency = "JPY"
)

type payload interface {
	Path() string
	Method() string
	GetOrderNo() string
	GetMerchantEmail() string
	GetTransferCurrency() PayInCurrency
}
type signer struct{}

func (signer) Sign(secret []byte, key string, payload payload) (string, string) {
	const (
		SignatureContent = "{ts}{method}{path}order_no={order_no}merchant_email={merchant_email}transfer_currency={transfer_currency}api_key={api_key}"
	)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	formater := strings.NewReplacer(
		"{ts}", ts,
		"{method}", payload.Method(),
		"{path}", payload.Path(),
		"{order_no}", payload.GetOrderNo(),
		"{merchant_email}", payload.GetMerchantEmail(),
		"{transfer_currency}", string(payload.GetTransferCurrency()),
		"{api_key}", key,
	)
	signature := formater.Replace(SignatureContent)
	hmac := hmac.New(sha256.New, secret)
	hmac.Write([]byte(signature))
	signature = hex.EncodeToString(hmac.Sum(nil))
	return ts, signature
}

type PayInRequest struct {
	MerchantOrderNo string
	RegisteredEmail string
	Amount          decimal.Decimal
	Currency        PayInCurrency
}

func (p *PayInRequest) toRaw(conf *Config) *rawPayInPayload {
	return &rawPayInPayload{
		OrderNo:          p.MerchantOrderNo,
		RegisteredEmail:  p.RegisteredEmail,
		TransferAmount:   p.Amount,
		TransferCurrency: p.Currency,

		MerchantEmail: conf.MerchantEmail,
		CallbackURL:   conf.CallbackURL,
		SuccessURL:    conf.SuccessURL,
		ReferrerURL:   conf.SuccessURL,
	}
}

// For JPY, it should be an integer. e.g. 100
// For other currencies, it will be rounded up to 2 decimal places. e.g. 100.00
type rawPayInPayload struct {
	MerchantEmail string `json:"merchant_email"`
	// generate by merchant,it should be unique.
	// The maximum number of characters is 127.
	OrderNo         string `json:"order_no"`
	RegisteredEmail string `json:"registered_email"` // peska user email.
	// Public IP address of the end-user. Ex: “215.81.64.12”
	ClientIP         string          `json:"client_ip,omitempty"`
	TransferAmount   decimal.Decimal `json:"transfer_amount"`
	TransferCurrency PayInCurrency   `json:"transfer_currency"`
	// Merchant callback url used for getting payment result.
	// This is the only callback url considered as source of truth from a merchant perspective.
	CallbackURL string `json:"callback_url,omitempty"`
	// After successful payment, a link to this URL is displayed at the bottom of the success page.
	// If no value is set, the default URL is used.
	SuccessURL string `json:"success_url,omitempty"`
	// Set for links displayed at the bottom of pages other than the Success page.
	// Clicking on this link will automatically cancel the Pay-in and take you to the linked page.
	// The Pay-in URL will also be invalid.
	ReferrerURL string `json:"referrer_url,omitempty"`
	Message     string `json:"message,omitempty"`
}

func (rawPayInPayload) Path() string {
	return "/v1/merchant/transfer"
}

func (rawPayInPayload) Method() string {
	return http.MethodPost
}

func (p *rawPayInPayload) GetOrderNo() string {
	return p.OrderNo
}
func (p *rawPayInPayload) GetMerchantEmail() string {
	return p.MerchantEmail
}
func (p *rawPayInPayload) GetTransferCurrency() PayInCurrency {
	return p.TransferCurrency
}

func (p *rawPayInPayload) GenerateSignedRequest(conf *Config) (*http.Request, error) {
	body := bytes.NewBuffer(nil)
	if err := json.NewEncoder(body).Encode(p); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(p.Method(), p.Path(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("AX-AUTHORIZE", conf.Key)
	// if p.ReferrerURL != "" {
	// 	req.Header.Set("Referer", p.ReferrerURL)
	// }
	ts, signature := signer{}.Sign(conf.Secret, conf.Key, p)
	req.Header.Set("AX-TIMESTAMP", ts)
	req.Header.Set("AX-SIGNATURE", signature)

	return req, nil
}

func (rawPayInPayload) Reply() rawResponseBody[PayInResponseData] {
	return rawResponseBody[PayInResponseData]{}
}

type rawResponseBody[T any] struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`    // Success Data
	Message json.RawMessage `json:"message"` // Message
	Code    ErrorCode       `json:"code"`    // Error codes

	data       *T
	messageStr string
}

func (r *rawResponseBody[T]) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	type alias rawResponseBody[T]
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	if a.Success && a.Code == ErrorCodeSucess {
		bodyData := new(T)
		if err := json.Unmarshal(a.Data, bodyData); err != nil {
			return err
		}
		a.data = bodyData
	}

	if a.Code != ErrorCodeValueInvalid {
		var msg string
		if err := json.Unmarshal(a.Message, &msg); err != nil {
			return err
		}
		a.messageStr = msg
	} else {
		if len(a.Message) != 0 && a.Message[0] == '"' {
			var msg string
			if err := json.Unmarshal(a.Message, &msg); err != nil {
				return err
			}
			a.messageStr = msg
		}
	}

	r.Success = a.Success
	r.Code = a.Code
	r.Data = a.Data
	r.Message = a.Message

	r.data = a.data
	r.messageStr = a.messageStr

	return nil
}

func (r *rawResponseBody[T]) IsSuccess() bool {
	return r.Success && r.Code == ErrorCodeSucess
}

func (r *rawResponseBody[T]) GetError() error {
	if r.IsSuccess() {
		return nil
	}
	var baseErr error
	switch r.Code {
	case ErrorCodeForbidden:
		baseErr = ErrForbidden
	case ErrorCodeInvalidContent:
		baseErr = ErrInvalidContent
	case ErrorCodeMissingHeader:
		baseErr = ErrMissingHeader
	case ErrorCodeInvalidTimestamp:
		baseErr = ErrInvalidTimestamp
	case ErrorCodeCurrencyNotSupport:
		baseErr = ErrCurrencyNotSupport
	case ErrorCodeInvalidTransferAmount:
		baseErr = ErrInvalidTransferAmount
	case ErrorCodeAuthFailed:
		baseErr = ErrAuthFailed
	case ErrorCodeSignatureFailed:
		baseErr = ErrSignatureFailed
	case ErrorCodeMerchantNotExist:
		baseErr = ErrMerchantNotExist
	case ErrorCodeUserNotExist:
		baseErr = ErrUserNotExist
	case ErrorCodeMerchantOrderRepeat:
		baseErr = ErrMerchantOrderRepeat
	case ErrorCodeMerchantOrderNotExist:
		baseErr = ErrMerchantOrderNotExist
	case ErrorCodeValueInvalid:
		baseErr = ErrValueInvalid
	}
	return fmt.Errorf("%w: %s", baseErr, r.messageStr)
}

func (r *rawResponseBody[T]) GetData() *T {
	return r.data
}

type PayInReply struct {
	data *PayInResponseData
}

func (reply *PayInReply) Status() string {
	return reply.data.Status
}

func (reply *PayInReply) TradeURL() string {
	return reply.data.TradeURL
}

type PayInResponseData struct {
	OrderNo                 string          `json:"order_no"`                  // Order No. specified by the merchant
	MerchantEmail           string          `json:"merchant_email"`            // Merchant’s email
	RegisteredEmail         string          `json:"registered_email"`          // Email of the user who made the pay-in
	RegisteredAccountNumber int             `json:"registered_account_number"` // Account Number(Peska wallet) of the user who made the pay-in
	RegisteredName          string          `json:"registered_name"`           // User’s name who made the pay-in
	TransferCurrency        string          `json:"transfer_currency"`         // Currency of Pay-in
	TransferAmount          decimal.Decimal `json:"transfer_amount"`           // Amount Pay-in by the user.
	Status                  string          `json:"status"`                    // Current status
	TradeURL                string          `json:"trade_url"`                 // URL for Pay-in
}

type GetPayInRecordPayload struct {
	OrderNo          string
	TransferCurrency PayInCurrency
}

func (p *GetPayInRecordPayload) toRaw(conf *Config) *rawGetPayInRecordPayload {
	return &rawGetPayInRecordPayload{
		OrderNo:          p.OrderNo,
		TransferCurrency: p.TransferCurrency,
		MerchantEmail:    conf.MerchantEmail,
	}
}

type rawGetPayInRecordPayload struct {
	MerchantEmail    string        `json:"merchant_email"`
	OrderNo          string        `json:"order_no"`
	TransferCurrency PayInCurrency `json:"transfer_currency"`
}

func (rawGetPayInRecordPayload) Path() string {
	return "/v1/merchant/query"
}

func (rawGetPayInRecordPayload) Method() string {
	return http.MethodPost
}

func (p *rawGetPayInRecordPayload) GetOrderNo() string {
	return p.OrderNo
}
func (p *rawGetPayInRecordPayload) GetMerchantEmail() string {
	return p.MerchantEmail
}
func (p *rawGetPayInRecordPayload) GetTransferCurrency() PayInCurrency {
	return p.TransferCurrency
}

func (p *rawGetPayInRecordPayload) GenerateSignedRequest(secret []byte, key string) (*http.Request, error) {
	body := bytes.NewBuffer(nil)
	if err := json.NewEncoder(body).Encode(p); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(p.Method(), p.Path(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("AX-AUTHORIZE", key)

	ts, signature := signer{}.Sign(secret, key, p)
	req.Header.Set("AX-TIMESTAMP", ts)
	req.Header.Set("AX-SIGNATURE", signature)

	return req, nil
}

func (rawGetPayInRecordPayload) Reply() rawResponseBody[PayInRecord] {
	return rawResponseBody[PayInRecord]{}
}

type PayInRecord struct {
	OrderNo                 string
	MerchantEmail           string
	RegisteredEmail         string
	RegisteredAccountNumber int
	RegisteredName          string
	TransferCurrency        string
	TransferAmount          decimal.Decimal
	FeeSide                 string
	Fee                     decimal.Decimal
	TotalAmount             decimal.Decimal
	Status                  PayInStatus
	ExpirationDate          time.Time
}

func (p *PayInRecord) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	var rawPayload struct {
		OrderNo                 string          `json:"order_no"`
		MerchantEmail           string          `json:"merchant_email"`
		RegisteredEmail         string          `json:"registered_email"`
		RegisteredAccountNumber int             `json:"registered_account_number"`
		RegisteredName          string          `json:"registered_name"`
		TransferCurrency        string          `json:"transfer_currency"`
		TransferAmount          decimal.Decimal `json:"transfer_amount"`
		FeeSide                 string          `json:"fee_side"`
		Fee                     decimal.Decimal `json:"fee"`
		TotalAmount             decimal.Decimal `json:"total_amount"`
		Status                  PayInStatus     `json:"status"`
		ExpirationDate          string          `json:"expiration_date"`
	}

	if err := json.Unmarshal(data, &rawPayload); err != nil {
		return err
	}

	tt, err := time.Parse(time.DateTime, rawPayload.ExpirationDate)
	if err != nil {
		return err
	}
	p.ExpirationDate = tt
	p.OrderNo = rawPayload.OrderNo
	p.MerchantEmail = rawPayload.MerchantEmail
	p.RegisteredEmail = rawPayload.RegisteredEmail
	p.RegisteredAccountNumber = rawPayload.RegisteredAccountNumber
	p.RegisteredName = rawPayload.RegisteredName
	p.TransferCurrency = rawPayload.TransferCurrency
	p.TransferAmount = rawPayload.TransferAmount
	p.FeeSide = rawPayload.FeeSide
	p.Fee = rawPayload.Fee
	p.TotalAmount = rawPayload.TotalAmount
	p.Status = rawPayload.Status

	return nil
}

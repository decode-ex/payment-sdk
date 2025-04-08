package peska

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type PayInStatus = string

const (
	PayInStatusPending   PayInStatus = "processing"
	PayInStatusCompleted PayInStatus = "process_complete"
	PayInStatusCanceled  PayInStatus = "cancel"
)

var ErrInvalidSign = errors.New("invalid sign")

type PayInCallbackPayload struct {
	OrderNo                 string
	MerchantEmail           string
	RegisteredEmail         string
	RegisteredAccountNumber int
	RegisteredName          string
	TransferCurrency        PayInCurrency
	TransferAmount          decimal.Decimal
	Fee                     decimal.Decimal
	TotalAmount             decimal.Decimal
	PayinID                 int
	Status                  PayInStatus
	TransferID              string
	CompletedAt             time.Time
	CancelReason            *string
	Message                 *string
	Signature               string
}

func (p *PayInCallbackPayload) IsCompleted() bool {
	return p.Status == PayInStatusCompleted
}

func (p *PayInCallbackPayload) IsCanceled() bool {
	return p.Status == PayInStatusCanceled
}

func (p *PayInCallbackPayload) VerifySignature(secret []byte, key string) error {
	const (
		SignatureContent = "{method}merchant_email={merchant_email}api_key={api_key}"
	)
	formater := strings.NewReplacer(
		"{method}", http.MethodPost,
		"{merchant_email}", p.MerchantEmail,
		"{api_key}", key,
	)
	signature := formater.Replace(SignatureContent)
	hmac := hmac.New(sha256.New, secret)
	hmac.Write([]byte(signature))
	signature = hex.EncodeToString(hmac.Sum(nil))
	if strings.EqualFold(p.Signature, signature) {
		return nil
	}
	return fmt.Errorf("%w, expect %s, got %s", ErrInvalidSign, signature, p.Signature)
}

func (p *PayInCallbackPayload) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	var raw struct {
		OrderNo                 string          `json:"order_no"`
		MerchantEmail           string          `json:"merchant_email"`
		RegisteredEmail         string          `json:"registered_email"`
		RegisteredAccountNumber int             `json:"registered_account_number"`
		RegisteredName          string          `json:"registered_name"`
		TransferCurrency        PayInCurrency   `json:"transfer_currency"`
		TransferAmount          decimal.Decimal `json:"transfer_amount"`
		Fee                     decimal.Decimal `json:"fee"`
		TotalAmount             decimal.Decimal `json:"total_amount"`
		PayinID                 int             `json:"payin_id"`
		Status                  PayInStatus     `json:"status"`
		TransferID              string          `json:"transfer_id"`
		CompletedAt             *string         `json:"completed_at"`
		CancelReason            *string         `json:"cancel_reason"`
		Message                 *string         `json:"message"`
		Signature               string          `json:"signature"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if at := raw.CompletedAt; at != nil {
		t, err := time.Parse(time.DateTime, *at)
		if err != nil {
			return err
		}
		p.CompletedAt = t
	}

	p.OrderNo = raw.OrderNo
	p.MerchantEmail = raw.MerchantEmail
	p.RegisteredEmail = raw.RegisteredEmail
	p.RegisteredAccountNumber = raw.RegisteredAccountNumber
	p.RegisteredName = raw.RegisteredName
	p.TransferCurrency = raw.TransferCurrency
	p.TransferAmount = raw.TransferAmount
	p.Fee = raw.Fee
	p.TotalAmount = raw.TotalAmount
	p.PayinID = raw.PayinID
	p.Status = raw.Status
	p.TransferID = raw.TransferID
	p.CancelReason = raw.CancelReason
	p.Message = raw.Message
	p.Signature = raw.Signature

	return nil
}

type PayInCallbackRequest struct {
	data *PayInCallbackPayload
}

func ParsePayInCallbackRequest(req *http.Request) (*PayInCallbackRequest, error) {
	if req.Method != http.MethodPost {
		return nil, fmt.Errorf("invalid method %s", req.Method)
	}

	var payload PayInCallbackPayload
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode request body failed: %w", err)
	}

	return &PayInCallbackRequest{data: &payload}, nil
}

func (req *PayInCallbackRequest) MerchantOrderNo() string {
	return req.data.OrderNo
}

func (req *PayInCallbackRequest) Amount() decimal.Decimal {
	return req.data.TotalAmount
}

func (req *PayInCallbackRequest) Currency() PayInCurrency {
	return req.data.TransferCurrency
}

func (req *PayInCallbackRequest) Status() PayInStatus {
	return req.data.Status
}

func (req *PayInCallbackRequest) SupplierOrderCode() string {
	return req.data.TransferID
}

func (req *PayInCallbackRequest) IsSuccess() bool {
	return req.data.IsCompleted()
}

func (req *PayInCallbackRequest) VerifySignature(conf *Config) error {
	if conf == nil {
		return fmt.Errorf("config is nil")
	}
	if req == nil || req.data == nil {
		return fmt.Errorf("payload is nil")
	}

	if len(conf.Secret) == 0 || len(conf.Key) == 0 {
		return fmt.Errorf("secret or key is empty")
	}
	if conf.MerchantEmail != req.data.MerchantEmail {
		return fmt.Errorf("merchant email not match")
	}

	return req.data.VerifySignature(conf.Secret, conf.Key)
}

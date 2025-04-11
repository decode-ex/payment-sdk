package asiabank

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/shopspring/decimal"
)

type PaymentStatus = string

const (
	// PENDING, indicate transaction is pending for further action from customer
	PaymentStatusPending PaymentStatus = "0"
	// SUCCESS, indicate transaction is accepted by gateway
	PaymentStatusSuccess PaymentStatus = "1"
	// FAIL, indicate transaction is rejected by gateway
	PaymentStatusFailed PaymentStatus = "2"
	// AUTHORIZED, used by Payout module. indicate transaction have been authorized.
	PaymentStatusAuthorized PaymentStatus = "3"
	// PROCESSING, used by Direct Debit and Payout module. indicate authorized transaction is processing by system.
	PaymentStatusProcessing PaymentStatus = "4"
)

type rawPaymentCallbackPayload struct {
	// same value from your payment request
	// string (36)
	MerchantReference string `form:"merchant_reference"`
	// unique reference assigned by payment gateway
	// string (36)
	RequestReference string `form:"request_reference"`
	// Currency ISO Code 4217, e.g. HKD, USD, CNY
	// string (3)
	Currency string `form:"currency"`
	// e.g. 10000.00,100.00, 1.00
	// double (11,2)
	Amount string `form:"amount"`
	// transaction status
	// string(2)
	Status string `form:"status"`
	// SHA-512 hashed signature
	// string (128)
	Sign string `form:"sign"`
}

func (p *rawPaymentCallbackPayload) generateSign(secret string) string {
	data := map[string]string{
		"merchant_reference": p.MerchantReference,
		"request_reference":  p.RequestReference,
		"currency":           p.Currency,
		"amount":             p.Amount,
		"status":             p.Status,
	}
	signer := signer{}
	return signer.Sign(data, secret)
}

func (payload *rawPaymentCallbackPayload) VerifySignature(secret string) error {
	sign := payload.generateSign(secret)

	if !strings.EqualFold(payload.Sign, sign) {
		return fmt.Errorf("signature not match, expect %s, got %s", sign, payload.Sign)
	}
	return nil
}

type PaymentCallbackRequest struct {
	data   *rawPaymentCallbackPayload
	amount decimal.Decimal
}

func (p *PaymentCallbackRequest) MerchantOrderID() string {
	return p.data.MerchantReference
}

func (req *PaymentCallbackRequest) Amount() decimal.Decimal {
	return req.amount
}

func (req *PaymentCallbackRequest) Currency() string {
	return req.data.Currency
}

func (req *PaymentCallbackRequest) Status() PaymentStatus {
	return req.data.Status
}

func (req *PaymentCallbackRequest) SupplierOrderCode() string {
	return req.data.RequestReference
}

func (req *PaymentCallbackRequest) VerifySignature(conf *Config) error {
	if conf == nil {
		return fmt.Errorf("config is nil")
	}
	if req == nil || req.data == nil {
		return fmt.Errorf("raw payload is nil")
	}

	return req.data.VerifySignature(conf.SecretKey)
}

func (req *PaymentCallbackRequest) IsSuccess() bool {
	return req.Status() == PaymentStatusSuccess
}

func ParsePaymentCallbackRequest(req *http.Request) (*PaymentCallbackRequest, error) {
	if req.Method != http.MethodPost {
		return nil, fmt.Errorf("invalid method: %s", req.Method)
	}
	if err := req.ParseForm(); err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}
	var payload rawPaymentCallbackPayload
	payload.Amount = req.FormValue("amount")
	payload.MerchantReference = req.FormValue("merchant_reference")
	payload.RequestReference = req.FormValue("request_reference")
	payload.Currency = req.FormValue("currency")
	payload.Status = req.FormValue("status")
	payload.Sign = req.FormValue("sign")

	amount, err := decimal.NewFromString(payload.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount: %w", err)
	}
	return &PaymentCallbackRequest{
		data:   &payload,
		amount: amount,
	}, nil
}

type PaymentCallbackReply struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    any    `json:"data"`
}

func (req *PaymentCallbackRequest) GenerateReply() *PaymentCallbackReply {
	return &PaymentCallbackReply{
		Code:    1,
		Message: "success",
		Success: true,
		Data:    nil,
	}
}

func (reply *PaymentCallbackReply) WriteTo(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(reply)
}

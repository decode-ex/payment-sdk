package bft

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/shopspring/decimal"
)

var ErrInvalidSign = errors.New("invalid sign")

type TradeStatus = string

const (
	TradeStatusSuccess TradeStatus = "1" // 成功
)

type rawCheckoutCallbackPayload struct {
	ApiOrderNo  string      `json:"apiOrderNo"`  // 商户订单号
	Money       string      `json:"money"`       // 订单金额
	TradeStatus TradeStatus `json:"tradeStatus"` // 交易状态。1：成功，其它为失败
	TradeID     string      `json:"tradeId"`     // Exlink订单号
	UniqueCode  string      `json:"uniqueCode"`  // 商户具有代表性的唯一标识
	Signature   string      `json:"signature"`   // 签名字符串
}

func (payload *rawCheckoutCallbackPayload) generateSignature(key string) string {
	signer := signer{}
	entries := []signEntry{
		{"apiOrderNo", payload.ApiOrderNo},
		{"money", payload.Money},
		{"tradeStatus", payload.TradeStatus},
		{"tradeId", payload.TradeID},
		{"uniqueCode", payload.UniqueCode},
	}
	signature := signer.Sign(key, entries...)
	return signature
}

func (payload *rawCheckoutCallbackPayload) VerifySignature(key string) error {

	expect := payload.generateSignature(key)
	if !strings.EqualFold(expect, payload.Signature) {
		return fmt.Errorf("%w, expect %s, got %s", ErrInvalidSign, expect, payload.Signature)
	}
	return nil
}

type CheckoutCallbackRequest struct {
	raw   *rawCheckoutCallbackPayload
	money decimal.Decimal
}

func ParseFundInCallbackRequest(req *http.Request) (*CheckoutCallbackRequest, error) {
	if req.Method != http.MethodPost {
		return nil, fmt.Errorf("invalid method: %s", req.Method)
	}
	var payload rawCheckoutCallbackPayload
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	money, err := decimal.NewFromString(payload.Money)
	if err != nil {
		return nil, fmt.Errorf("invalid money: %w", err)
	}

	return &CheckoutCallbackRequest{
		raw:   &payload,
		money: money,
	}, nil
}

func (req *CheckoutCallbackRequest) MerchantOrderID() string {
	return req.raw.ApiOrderNo
}

func (req *CheckoutCallbackRequest) Amount() decimal.Decimal {
	return req.money
}

func (req *CheckoutCallbackRequest) Currency() string {
	return "CNY"
}

func (req *CheckoutCallbackRequest) Status() TradeStatus {
	return req.raw.TradeStatus
}

func (req *CheckoutCallbackRequest) SupplierOrderCode() string {
	return req.raw.TradeID
}

func (req *CheckoutCallbackRequest) VerifySignature(conf *Config) error {
	if conf == nil {
		return fmt.Errorf("config is nil")
	}
	if req == nil || req.raw == nil {
		return fmt.Errorf("raw payload is nil")
	}

	return req.raw.VerifySignature(conf.PublicKey)
}

func (req *CheckoutCallbackRequest) IsSuccess() bool {
	return req.Status() == TradeStatusSuccess
}

type CheckoutCallbackReply struct {
	Code    responseCode `json:"code"`    // 交易状态。1：成功，其它为失败
	Message string       `json:"message"` // 结果说明
	Data    any          `json:"data"`    // 接口返回结果
	Success bool         `json:"success"` // true:成功，false:失败
}

func (req *CheckoutCallbackRequest) Reply() *CheckoutCallbackReply {
	return &CheckoutCallbackReply{
		Code:    responseCodeSuccess,
		Message: "success",
		Data:    nil,
		Success: true,
	}
}

func (reply *CheckoutCallbackReply) Write(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(reply)
}

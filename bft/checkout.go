package bft

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/decode-ex/payment-sdk/internal/strings2"
	"github.com/shopspring/decimal"
)

var ErrorInvalidData = errors.New("invalid data")

type signEntry struct {
	Key   string
	Value string
}
type signer struct{}

func (signer) Sign(privateKey string, entries ...signEntry) string {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})

	var signString strings.Builder
	for _, entry := range entries {
		signString.WriteString(entry.Key)
		signString.WriteString("=")
		signString.WriteString(entry.Value)
		signString.WriteString("&")
	}
	signString.WriteString("key=")
	signString.WriteString(privateKey)

	content := signString.String()
	bs := md5.Sum(strings2.ToBytesNoAlloc(content))
	return hex.EncodeToString(bs[:])
}

type CheckoutRequest struct {
	CustomerID      string
	Amount          decimal.Decimal
	MerchantOrderID string
	CustomerName    string
}

func (req *CheckoutRequest) Validate() error {
	if req.CustomerID == "" {
		return ErrorInvalidData
	}
	if req.MerchantOrderID == "" {
		return ErrorInvalidData
	}
	if req.CustomerName == "" {
		return ErrorInvalidData
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return ErrorInvalidData
	}

	amount := req.Amount.Truncate(0)
	if !amount.Equal(req.Amount) {
		return ErrorInvalidData
	}

	return nil
}

func (req *CheckoutRequest) toRaw(config *Config) *rawCheckoutPayload {
	raw := &rawCheckoutPayload{
		Uid:        config.MerchantID,
		UniqueCode: req.CustomerID,
		Money:      req.Amount.StringFixed(0),
		PayType:    config.DefaultPayType,
		OrderID:    req.MerchantOrderID,
		PayerName:  req.CustomerName,
	}
	raw.Signature = raw.GenrateSignature(config.PrivateKey)
	return raw
}

type PayType = string

const (
	PayTypeUnionPay PayType = "1" // 银联
)

type rawCheckoutPayload struct {
	// 商户UID,对应商户后台的“商户编码"
	Uid string `json:"uid"`
	// 商户具有代表性的唯一标识。例如：用户ID，业务ID等
	UniqueCode string `json:"uniqueCode"`
	// 金额为整数。单位：人民币
	Money string `json:"money"`
	// 支付类型。1：银联
	PayType PayType `json:"payType"`
	// 商户订单号。商户平台自己生成的单号
	OrderID string `json:"orderId"`
	// 付款人名字
	PayerName string `json:"payerName"`
	// 签名字符串
	Signature string `json:"signature"`
}

func (payload rawCheckoutPayload) Path() string {
	return "/coin/pay/order/pay/checkout/counter"
}

func (payload rawCheckoutPayload) Method() string {
	return http.MethodPost
}

func (payload *rawCheckoutPayload) GenrateSignature(key string) string {
	signer := signer{}
	entries := []signEntry{
		{"uid", payload.Uid},
		{"uniqueCode", payload.UniqueCode},
		{"money", payload.Money},
		{"payType", payload.PayType},
		{"orderId", payload.OrderID},
		{"payerName", payload.PayerName},
	}
	signature := signer.Sign(key, entries...)
	return signature
}

func (payload *rawCheckoutPayload) GenerateSignedRequest(ctx context.Context, config *Config) (*http.Request, error) {
	const (
		Path        = "/coin/pay/order/pay/checkout/counter"
		Method      = http.MethodPost
		ContentType = "application/json"
	)

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, Method, Path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}
	req.Header.Set("Content-Type", ContentType)

	return req, nil
}

func (rawCheckoutPayload) Reply() *rawCheckoutResponse {
	return &rawCheckoutResponse{}
}

type responseCode = int

const (
	responseCodeSuccess responseCode = 1
)

type rawCheckoutResponse struct {
	// 接口调用状态，1:成功，其他值：失败
	Code responseCode `json:"code"`
	// 结果说明，如果接口调用出错，那么返回错误描述，成功返回“成功”
	Message string `json:"message"`
	// 接口返回结果。值为URL地址，拿到这个URL可以跳转到下单页面
	Data string `json:"data"`
	// true:成功，false:失败
	Success bool `json:"success"`
}

type CheckoutReply struct {
	RedirectURL string
}

func (CheckoutReply) fromRaw(raw *rawCheckoutResponse) (*CheckoutReply, error) {
	if raw == nil {
		return nil, fmt.Errorf("raw response is nil")
	}
	if raw.Code != responseCodeSuccess || !raw.Success {
		return nil, fmt.Errorf("checkout failed %s", raw.Message)
	}
	return &CheckoutReply{
		RedirectURL: raw.Data,
	}, nil
}

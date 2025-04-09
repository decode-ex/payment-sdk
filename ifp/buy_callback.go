package ifp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrInvalidSign = errors.New("invalid sign")
)

type rawBuyCallbackPayload struct {
	Success    bool
	StatusCode IFPStatusCode
	Message    string
	Signature  string // 需要通过鉴权算法进行校验
	Time       time.Time
	Data       rawBuyCallbackPayloadData `json:"data"`
}

func (req *rawBuyCallbackPayload) IsSuccess() bool {
	return req.Success && req.StatusCode == IFPStatusCode_Success
}

func (req *rawBuyCallbackPayload) Validate() error {
	if req.StatusCode == "" {
		return errors.New("missing statusCode")
	}
	if req.Signature == "" {
		return errors.New("missing signature")
	}
	return req.Data.Validate()
}

func (req *rawBuyCallbackPayload) UnmarshalJSON(content []byte) error {
	if bytes.Equal(content, []byte("null")) {
		return nil
	}
	if req == nil {
		panic("unreachable")
	}

	var temp struct {
		Success    bool                       `json:"success"`
		StatusCode IFPStatusCode              `json:"statusCode"`
		Message    string                     `json:"message"`
		Signature  string                     `json:"signature"` // 需要通过鉴权算法进行校验
		Timestamp  int64                      `json:"timestamp"` // UTC 毫秒级时间戳
		Data       *rawBuyCallbackPayloadData `json:"data"`
	}
	if err := json.Unmarshal(content, &temp); err != nil {
		return err
	}
	if temp.Timestamp == 0 {
		return errors.New("missing timestamp")
	}
	if temp.Data == nil {
		return errors.New("missing data")
	}

	req.Success = temp.Success
	req.StatusCode = temp.StatusCode
	req.Message = temp.Message
	req.Signature = temp.Signature
	req.Time = time.UnixMilli(temp.Timestamp)
	req.Data = *temp.Data
	return nil
}

type rawBuyCallbackPayloadData struct {
	Ticket                string
	TransactionCode       string          // 订单编码
	TransactionAmount     decimal.Decimal // 交易USDD数量
	CurrencyCode          CurrencyCode    // 支付币种
	PaymentPrice          decimal.Decimal // 支付金额
	TransactionCreateTime time.Time       // UTC 交易创建时间
	PaymentFinishedTime   time.Time       // UTC 支付完成时间
}

func (data *rawBuyCallbackPayloadData) Validate() error {
	if data.Ticket == "" {
		return errors.New("missing externalOrderNumber")
	}
	if data.TransactionCode == "" {
		return errors.New("missing transactionCode")
	}
	return nil
}

func (data *rawBuyCallbackPayloadData) UnmarshalJSON(content []byte) error {
	if bytes.Equal(content, []byte("null")) {
		return nil
	}
	if data == nil {
		panic("unreachable")
	}

	var temp struct {
		ExternalOrderNumber   string `json:"externalOrderNumber"`   // 外部订单号, 对应平台流水ticket
		TransactionCode       string `json:"transactionCode"`       // 订单编号, 即三方平台订单号
		TransactionAmount     string `json:"transactionAmount"`     // 交易USDD数量
		CurrencyCode          string `json:"currencyCode"`          // 支付币种
		PaymentPrice          string `json:"paymentPrice"`          // 支付金额
		TransactionCreateTime string `json:"transactionCreateTime"` // UTC 交易创建时间
		PaymentFinishedTime   string `json:"paymentFinishedTime"`   // UTC 支付完成时间
	}
	if err := json.Unmarshal(content, &temp); err != nil {
		return err
	}
	if temp.TransactionCreateTime == "" {
		return errors.New("missing transactionCreateTime")
	}

	if temp.TransactionAmount != "" {
		t, err := decimal.NewFromString(temp.TransactionAmount)
		if err != nil {
			return fmt.Errorf("invalid transactionAmount %s %w", temp.TransactionAmount, err)
		}
		data.TransactionAmount = t
	}

	if temp.PaymentPrice != "" {
		t, err := decimal.NewFromString(temp.PaymentPrice)
		if err != nil {
			return fmt.Errorf("invalid paymentPrice %s %w", temp.PaymentPrice, err)
		}
		data.PaymentPrice = t
	}

	data.Ticket = temp.ExternalOrderNumber
	data.TransactionCode = temp.TransactionCode
	data.CurrencyCode = CurrencyCode(temp.CurrencyCode)

	transactionCreateTime, err := time.ParseInLocation(time.DateTime, temp.TransactionCreateTime, time.UTC)
	if err != nil {
		return fmt.Errorf("invalid transactionCreateTime %s %w", temp.TransactionCreateTime, err)
	}
	data.TransactionCreateTime = transactionCreateTime

	if temp.PaymentFinishedTime != "" {
		paymentFinishedTime, err := time.ParseInLocation(time.DateTime, temp.PaymentFinishedTime, time.UTC)
		if err != nil {
			return fmt.Errorf("invalid paymentFinishedTime %s %w", temp.PaymentFinishedTime, err)
		}
		data.PaymentFinishedTime = paymentFinishedTime
	}
	return nil
}

type BuyCallbackRequest struct {
	*baseRequest
	payload *rawBuyCallbackPayload
}

func ParseBuyCallbackRequest(req *http.Request) (*BuyCallbackRequest, error) {
	if req.Method != http.MethodPost {
		return nil, fmt.Errorf("invalid method %s", req.Method)
	}
	var payload rawBuyCallbackPayload
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("json decode error: %w", err)
	}
	if err := payload.Validate(); err != nil {
		return nil, fmt.Errorf("payload validate error: %w", err)
	}
	return &BuyCallbackRequest{
		baseRequest: newBaseRequestWithTimestamp(payload.Time.UnixMilli()),
		payload:     &payload,
	}, nil
}

func (req *BuyCallbackRequest) MerchantOrderID() string {
	return req.payload.Data.Ticket
}
func (req *BuyCallbackRequest) Amount() decimal.Decimal {
	return req.payload.Data.TransactionAmount
}
func (req *BuyCallbackRequest) Currency() CurrencyCode {
	return req.payload.Data.CurrencyCode
}
func (req *BuyCallbackRequest) Status() IFPStatusCode {
	return req.payload.StatusCode
}
func (req *BuyCallbackRequest) SupplierOrderCode() string {
	return req.payload.Data.TransactionCode
}
func (req *BuyCallbackRequest) VerifySignature(conf *Config) error {
	if conf == nil {
		return fmt.Errorf("config is nil")
	}
	if req == nil || req.payload == nil {
		return fmt.Errorf("payload is nil")
	}

	signature := req.GenerateSignature(conf.AccessKey, conf.PrivateKey)
	if strings.EqualFold(signature, req.payload.Signature) {
		return nil
	}
	return fmt.Errorf("invalid signature, expect %s, got %s", signature, req.payload.Signature)
}

func (req *BuyCallbackRequest) IsSuccess() bool {
	return req.payload.IsSuccess()
}

type BuyCallbackReply struct{}

func (req *BuyCallbackRequest) GenerateReply() *BuyCallbackReply {
	return &BuyCallbackReply{}
}

func (reply *BuyCallbackReply) encode() string {
	return `{"success":true}`
}

func (reply *BuyCallbackReply) WriteTo(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(reply.encode()))
	return err
}

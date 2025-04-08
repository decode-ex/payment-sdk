package ifp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/shopspring/decimal"
)

type OrderStatus = int

// 0 => created, 1 => fiat_transfered, 2 => confirmed, 3 => canceled, 4 => confirmed_by_admin, 5 => discarded_by_admin
// 6 => 超时未支付取消(文档未记录, 测试得到, 超时时间25分钟)
const (
	OrderStatus_Created OrderStatus = iota
	OrderStatus_FiatTransfered
	OrderStatus_Confirmed
	OrderStatus_Canceled
	OrderStatus_ConfirmedByAdmin
	OrderStatus_DiscardedByAdmin
	OrderStatus_TimeoutCanceled
)

type OrderInfo struct {
	CallbackURL           string          `json:"callbackUrl"`
	ID                    string          `json:"code"`
	CurrencyCode          CurrencyCode    `json:"currencyCode"`
	PayerRealName         string          `json:"payerRealName"`
	PaymentFinishedTime   time.Time       `json:"paymentFinishedTime"`
	Status                OrderStatus     `json:"status"`
	TotalPrice            decimal.Decimal `json:"totalPrice"`
	TransactionCreateTime time.Time       `json:"transactionCreateTime"`
	UnitPrice             decimal.Decimal `json:"unitPrice"`
	UsddAmount            decimal.Decimal `json:"usddAmount"`
}

type QueryOrderResponse = IFPGenericResponse[OrderInfo]

func (oi *OrderInfo) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}
	if oi == nil {
		panic("unreachable")
	}

	var tmp struct {
		CallbackURL           string `json:"callbackUrl"`
		Code                  string `json:"code"`
		CurrencyCode          string `json:"currencyCode"`
		PayerRealName         string `json:"payerRealName"`
		PaymentFinishedTime   string `json:"paymentFinishedTime"`
		Status                int    `json:"status"`
		TotalPrice            string `json:"totalPrice"`
		TransactionCreateTime string `json:"transactionCreateTime"`
		UnitPrice             string `json:"unitPrice"`
		UsddAmount            string `json:"usddAmount"`
	}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	if tmp.TransactionCreateTime == "" {
		return errors.New("missing transactionCreateTime")
	}

	t, err := time.ParseInLocation(time.DateTime, tmp.TransactionCreateTime, time.UTC)
	if err != nil {
		return err
	}
	oi.TransactionCreateTime = t

	if tmp.PaymentFinishedTime != "" {
		t, err := time.ParseInLocation(time.DateTime, tmp.PaymentFinishedTime, time.UTC)
		if err != nil {
			return err
		}
		oi.PaymentFinishedTime = t
	}

	tp, err := decimal.NewFromString(tmp.TotalPrice)
	if err != nil {
		return err
	}
	oi.TotalPrice = tp

	up, err := decimal.NewFromString(tmp.UnitPrice)
	if err != nil {
		return err
	}
	oi.UnitPrice = up

	ua, err := decimal.NewFromString(tmp.UsddAmount)
	if err != nil {
		return err
	}
	oi.UsddAmount = ua

	oi.CallbackURL = tmp.CallbackURL
	oi.ID = tmp.Code
	oi.CurrencyCode = CurrencyCode(tmp.CurrencyCode)
	oi.PayerRealName = tmp.PayerRealName
	oi.Status = OrderStatus(tmp.Status)

	return nil
}

type rawQueryOrderRequest struct {
	baseRequest
	ExternalOrderNumber string `json:"externalOrderNumber"`
}

type QueryOrderRequest struct {
	MerchantOrderID string
}

func (req *QueryOrderRequest) Validate() error {
	if req.MerchantOrderID == "" {
		return fmt.Errorf("merchant order id is empty")
	}
	return nil
}

func (qr *QueryOrderRequest) toRaw(_ *Config) *rawQueryOrderRequest {
	return &rawQueryOrderRequest{
		ExternalOrderNumber: qr.MerchantOrderID,
	}
}

func (raw *rawQueryOrderRequest) GenerateSignedRequest(ctx context.Context, conf *Config) (*http.Request, error) {
	const (
		Endpoint_QueryOrder = "/api/get-order"
		Method              = http.MethodGet
	)

	uri, _ := url.JoinPath(Endpoint_QueryOrder, raw.ExternalOrderNumber)
	req, _ := http.NewRequestWithContext(ctx, Method, uri, nil)

	req.Header.Set(IFPHeaderKey_Accesskey, conf.AccessKey)
	req.Header.Set(IFPHeaderKey_Timestamp, raw.ts)
	req.Header.Set(IFPHeaderKey_Signature, raw.GenerateSignature(conf.AccessKey, conf.PrivateKey))
	return req, nil
}

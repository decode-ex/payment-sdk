package asiabank

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/shopspring/decimal"
)

const (
	_ENDPOINT_TEMPLATE = "https://payment.pa-sys.com/app/page/{MerchantToken}"
)

type Config struct {
	MerchantToken string

	SecretKey string

	SuccessURL string
}

type Client struct {
	config   *Config
	endpoint string
}

func NewClient(config Config) (*Client, error) {
	sr := strings.ReplaceAll(_ENDPOINT_TEMPLATE, "{MerchantToken}", config.MerchantToken)

	return &Client{
		config:   &config,
		endpoint: sr,
	}, nil
}

type PaymentRequest struct {
	MerchantOrderID string
	Currency        string
	Amount          decimal.Decimal

	CustomerIP        string
	CustomerFirstName string
	CustomerLastName  string
	CustomerPhone     string
	CustomerEmail     string
	CustomerCountry   string

	Network string
}

func (req *PaymentRequest) Validate() error {
	if req.MerchantOrderID == "" {
		return ErrInvalidMerchantOrderID
	}
	if req.Currency == "" {
		return ErrInvalidCurrency
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}
	if req.CustomerIP == "" {
		return ErrInvalidCustomerIP
	}
	if req.CustomerFirstName == "" {
		return ErrInvalidCustomerFirstName
	}
	if req.CustomerLastName == "" {
		return ErrInvalidCustomerLastName
	}
	if req.CustomerPhone == "" {
		return ErrInvalidCustomerPhone
	}
	if req.CustomerEmail == "" {
		return ErrInvalidCustomerEmail
	}
	if req.Network == "" {
		return ErrInvalidNetwork
	}
	return nil
}

func (req *PaymentRequest) toRaw(conf *Config) *rawPaymentForm {
	raw := &rawPaymentForm{
		MerchantReference: req.MerchantOrderID,
		Currency:          req.Currency,
		Amount:            req.Amount.StringFixedBank(2),
		ReturnURL:         conf.SuccessURL,
		CustomeIP:         req.CustomerIP,
		CustomerFirstName: req.CustomerFirstName,
		CustomerLastName:  req.CustomerLastName,
		CustomerPhone:     req.CustomerPhone,
		CustomerEmail:     req.CustomerEmail,
		CustomerCountry:   req.CustomerCountry,
		Network:           req.Network,
	}
	params := raw.toParams()
	raw.Sign = signer{}.Sign(params, conf.SecretKey)
	return raw
}

type PaymentForm struct {
	Method string
	Action string
	Fields url.Values
}

func (cli *Client) MakePaymentForm(ctx context.Context, req *PaymentRequest) (*PaymentForm, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	raw := req.toRaw(cli.config)
	form := &PaymentForm{
		Method: http.MethodPost,
		Action: cli.endpoint,
		Fields: raw.Encode(),
	}
	return form, nil
}

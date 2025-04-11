package asiabank

import (
	"context"
	"errors"
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

func NewClient(config Config) *Client {
	sr := strings.ReplaceAll(_ENDPOINT_TEMPLATE, "{MerchantToken}", config.MerchantToken)

	return &Client{
		config:   &config,
		endpoint: sr,
	}
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

	Network string
}

func (req *PaymentRequest) Validate() error {
	if req.MerchantOrderID == "" {
		return errors.New("merchant order id is empty")
	}
	if req.Currency == "" {
		return errors.New("currency is empty")
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount is invalid")
	}
	if req.CustomerIP == "" {
		return errors.New("customer ip is empty")
	}
	if req.CustomerFirstName == "" {
		return errors.New("customer first name is empty")
	}
	if req.CustomerLastName == "" {
		return errors.New("customer last name is empty")
	}
	if req.CustomerPhone == "" {
		return errors.New("customer phone is empty")
	}
	if req.CustomerEmail == "" {
		return errors.New("customer email is empty")
	}
	if req.Network == "" {
		return errors.New("network is empty")
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

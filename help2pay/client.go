package help2pay

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/text/language"
)

var (
	_DEV_ENDPOINT, _  = url.Parse("https://api.testingzone88.com")
	_PROD_ENDPOINT, _ = url.Parse("https://api.safepaymentapp.com")
)

type Env int

const (
	EnvDev Env = iota
	EnvProd
)

func (e Env) baseURL() *url.URL {
	switch e {
	case EnvDev:
		return _DEV_ENDPOINT
	case EnvProd:
		return _PROD_ENDPOINT
	default:
		return _DEV_ENDPOINT
	}
}

type Config struct {
	MerchantCode string
	SecurityCode string

	CompanyName string

	SuccessURL  string
	CallbackURL string

	tz *time.Location
}

type Client struct {
	env  Env
	conf *Config
}

func NewClient(env Env, conf Config) (*Client, error) {
	tz, err := time.LoadLocation("Asia/Chongqing")
	if err != nil {
		return nil, err
	}
	conf.tz = tz

	return &Client{
		conf: &conf,
		env:  env,
	}, nil
}

func NewDevClient(conf Config) (*Client, error) {
	return NewClient(EnvDev, conf)
}

func NewProdClient(conf Config) (*Client, error) {
	return NewClient(EnvProd, conf)
}

type DepositFormRequest struct {
	MerchantOrerID string

	Bank     string
	Currency CurrencyCode
	Amount   decimal.Decimal

	CustomerID string
	CustomerIP string
	Language   language.Tag
}

func (req *DepositFormRequest) Validate() error {
	if req.MerchantOrerID == "" {
		return ErrInvalidMerchantOrderID
	}
	if !IsCurrencySupportBank(req.Currency, req.Bank) {
		return errors.New("invalid bank or currency")
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}
	tr := req.Amount.Truncate(0)
	if req.Currency == CurrencyCodeVND || req.Currency == CurrencyCodeIDR {
		// VND, IDR currency and PPTP (THB currency) Will
		// Only Allow .00 decimal submission
		if !tr.Equal(req.Amount) {
			return ErrInvalidAmount
		}
	}
	if req.CustomerID == "" {
		return ErrInvalidCustomerID
	}
	if req.CustomerIP == "" {
		return ErrInvalidCustomerIP
	}
	return nil
}

func (req *DepositFormRequest) toRaw(conf *Config) *rawDepositFormRequest {
	raw := &rawDepositFormRequest{
		Merchant:    conf.MerchantCode,
		Currency:    req.Currency,
		Customer:    req.CustomerID,
		Reference:   req.MerchantOrerID,
		Amount:      req.Amount.StringFixedBank(2),
		Datetime:    req.formatDatetime(conf.tz),
		FrontURI:    conf.SuccessURL,
		BackURI:     conf.CallbackURL,
		Bank:        req.Bank,
		Language:    getLanguageCode(req.Language),
		ClientIP:    req.CustomerIP,
		CompanyName: conf.CompanyName,
	}
	raw.Key = signer{}.SignRequest(raw, conf.SecurityCode)
	return raw
}

func (req *DepositFormRequest) formatDatetime(tz *time.Location) time.Time {
	return time.Now().In(tz)
}

type DepositForm struct {
	Method string
	Action string
	Fields url.Values
}

func (cli *Client) MakeFiatDepositForm(_ context.Context, req *DepositFormRequest) (*DepositForm, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	raw := req.toRaw(cli.conf)

	return &DepositForm{
		Method: "POST",
		Action: cli.env.baseURL().ResolveReference(raw.Path()).String(),
		Fields: raw.Encode(),
	}, nil
}

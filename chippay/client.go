package chippay

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	httptransport "github.com/decode-ex/payment-sdk/internal/http_transport"
	"github.com/shopspring/decimal"
)

const (
	_DEV_BASE_URL  = "https://open-v2.chippaytest.com"
	_PROD_BASE_URL = "https://open-v2.chippay.com"
)

type Env int

const (
	EnvDev Env = iota
	EnvProd
)

func (e Env) baseURL() string {
	switch e {
	case EnvDev:
		return _DEV_BASE_URL
	case EnvProd:
		return _PROD_BASE_URL
	default:
		return _DEV_BASE_URL
	}
}

type Config struct {
	MerchantID string
	PublicKey  string
	PrivateKey string

	CallbackURL string
	RedirectURL string

	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

type Client struct {
	http   *http.Client
	config *Config
}

func NewClient(env Env, config Config) (*Client, error) {
	transport, err := httptransport.NewTransport(env.baseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	priKeyBytes, err := base64.StdEncoding.DecodeString(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}
	priKey, err := x509.ParsePKCS8PrivateKey(priKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	rsaPriKey, ok := priKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("invalid private key type")
	}

	pubKeyBytes, err := base64.StdEncoding.DecodeString(config.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	pubKey, err := x509.ParsePKIXPublicKey(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("invalid public key type")
	}

	return &Client{
		http: &http.Client{
			Transport: transport,
		},
		config: &Config{
			MerchantID:  config.MerchantID,
			PublicKey:   config.PublicKey,
			PrivateKey:  config.PrivateKey,
			CallbackURL: config.CallbackURL,
			RedirectURL: config.RedirectURL,
			privateKey:  rsaPriKey,
			publicKey:   rsaPubKey,
		},
	}, nil
}

type BuyCoinRequest struct {
	MerchantOrderID string

	Amount   decimal.Decimal
	Currency string

	CustomerPhone string
	CustomerName  string
}

func (raw *BuyCoinRequest) Validate() error {
	if raw.MerchantOrderID == "" {
		return errors.New("merchant order ID is required")
	}

	if raw.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be greater than zero")
	}

	amount := raw.Amount.Truncate(0)
	if !amount.Equal(raw.Amount) {
		return errors.New("amount must be an integer")
	}

	if raw.Currency == "" {
		return errors.New("currency is required")
	}

	currencyAllow := false
	for _, allow := range []string{"CNY", "VND"} {
		if strings.EqualFold(raw.Currency, allow) {
			currencyAllow = true
			break
		}
	}
	if !currencyAllow {
		return fmt.Errorf("currency %s is not supported", raw.Currency)
	}

	if raw.CustomerPhone == "" {
		return errors.New("customer phone is required")
	}
	if raw.CustomerName == "" {
		return errors.New("customer name is required")
	}
	return nil

}

func (req *BuyCoinRequest) toRaw(conf *Config) *rawBuyPayload {
	return &rawBuyPayload{
		CompanyID:       conf.MerchantID,
		KYCLevel:        "2",
		UserName:        req.CustomerName,
		Phone:           req.CustomerPhone,
		OrderType:       OrderTypeBuy,
		CompanyOrderNum: req.MerchantOrderID,
		CoinSign:        CoinSignUSDT,
		PayCoinSign:     strings.ToLower(req.Currency),
		Total:           req.Amount.StringFixed(0),
		OrderTime:       time.Now(),
		SyncURL:         conf.RedirectURL,
		AsyncUrl:        conf.CallbackURL,
	}
}

type BuyCoinReply struct {
	SupplyOrderNum string
	RedirectURL    string
}

func (BuyCoinReply) fromRaw(raw *rawBuyResponse) (*BuyCoinReply, error) {
	if raw == nil {
		return nil, errors.New("raw response is nil")
	}
	if raw.Code != StatusCodeSuccess {
		return nil, fmt.Errorf("failed to buy coin: %s", raw.Message)
	}
	return &BuyCoinReply{
		SupplyOrderNum: raw.Data.OrderNo,
		RedirectURL:    raw.Data.Link,
	}, nil
}

func (c *Client) BuyCoin(ctx context.Context, req *BuyCoinRequest) (*BuyCoinReply, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	raw := req.toRaw(c.config)
	buyReq, err := raw.GenerateSignedRequest(ctx, c.config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signed request: %w", err)
	}
	resp, err := c.http.Do(buyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	rawReply := raw.Reply()
	if err := json.NewDecoder(resp.Body).Decode(&rawReply); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return BuyCoinReply{}.fromRaw(&rawReply)
}

package long77

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	httptransport "github.com/decode-ex/payment-sdk/internal/http_transport"
)

const (
	_DEV_BASE_URL  = "https://test.ddtpay.org/"
	_PROD_BASE_URL = "https://ddtpay.org/"
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

type Client struct {
	http   *http.Client
	config *Config
}

type Config struct {
	NotifyURL string // Callback url
	ReturnURL string // Success url

	PartnerID string //
	Secret    string //
}

func NewClient(env Env, conf Config) (*Client, error) {
	transport, err := httptransport.NewTransport(env.baseURL())
	if err != nil {
		return nil, err
	}

	return &Client{
		http: &http.Client{
			Transport: transport,
		},
		config: &conf,
	}, nil
}

func NewDevClient(conf Config) (*Client, error) {
	return NewClient(EnvDev, conf)
}

func NewProdClient(conf Config) (*Client, error) {
	return NewClient(EnvProd, conf)
}

func (c *Client) CreatePayInURL(ctx context.Context, in *PayInRequest) (*PayInResponse, error) {
	raw, err := in.toRaw(c.config)
	if err != nil {
		return nil, err
	}
	req, err := raw.GenerateSignedRequest(c.config)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var out rawPayInResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.toResponse()
}

package bft

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	httptransport "github.com/decode-ex/payment-sdk/internal/http_transport"
)

const (
	_DEV_BASE_URL  = "https://api.maxpay666.com"
	_PROD_BASE_URL = "https://api.exlinked.com"
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
	MerchantID     string
	DefaultPayType PayType
	PublicKey      string
	PrivateKey     string
}

type Client struct {
	http   *http.Client
	config *Config
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

func (cli *Client) CheckoutOut(ctx context.Context, req *CheckoutRequest) (*CheckoutReply, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	raw := req.toRaw(cli.config)
	reqBody, err := raw.GenerateSignedRequest(ctx, cli.config)
	if err != nil {
		return nil, err
	}
	resp, err := cli.http.Do(reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	reply := raw.Reply()
	if err := json.NewDecoder(resp.Body).Decode(reply); err != nil {
		return nil, err
	}

	return CheckoutReply{}.fromRaw(reply)
}

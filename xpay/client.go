package xpay

import (
	"context"
	"net/url"
)

type Client struct {
	baseURL *url.URL
	conf    *Config
}

type Config struct {
	// xpay base url
	BaseURL string
	// webhook callback url
	CallbackURL string
	// success redirect url
	SuccessURL string
	// merchant id
	MerchantID string
	// secret key
	Key string
}

func NewClient(conf Config) (*Client, error) {
	base, err := url.Parse(conf.BaseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		baseURL: base,
		conf:    &conf,
	}, nil
}

func (cli *Client) CreateFundInURL(ctx context.Context, req *FundInRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", err
	}

	fundInReq, err := req.toRaw(cli.conf).GenerateSignedRequest(ctx, cli.conf)
	if err != nil {
		return "", err
	}

	url := cli.baseURL.ResolveReference(fundInReq.URL)
	url.RawQuery = fundInReq.URL.RawQuery
	url.RawFragment = fundInReq.URL.RawFragment
	return url.String(), nil
}

package xpay

import (
	"context"
	"net/url"
)

var (
	_BASE_URL, _ = url.Parse("https://bo.transfer1515.com/")
)

type Client struct {
	conf *Config
}

type Config struct {
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
	return &Client{
		conf: &conf,
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

	url := _BASE_URL.ResolveReference(fundInReq.URL)
	url.RawQuery = fundInReq.URL.RawQuery
	url.RawFragment = fundInReq.URL.RawFragment
	return url.String(), nil
}

package peska

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	http *http.Client
	conf *Config
}

type Config struct {
	BaseURL       string
	CallbackURL   string
	SuccessURL    string
	MerchantEmail string

	Secret []byte
	Key    string
}

type transport struct {
	inner   http.RoundTripper
	baseURL *url.URL
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	uri := t.baseURL.ResolveReference(req.URL)
	req.URL = uri

	req.Header.Add("Referer", t.baseURL.String())
	return t.inner.RoundTrip(req)
}

func NewClient(conf Config) (*Client, error) {
	base, err := url.Parse(conf.BaseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		http: &http.Client{
			Transport: &transport{
				inner:   http.DefaultTransport,
				baseURL: base,
			},
		},
		conf: &conf,
	}, nil
}

func (cli *Client) CreatePayInURL(ctx context.Context, payload *PayInRequest) (*PayInReply, error) {
	raw := payload.toRaw(cli.conf)
	req, err := raw.GenerateSignedRequest(cli.conf)
	if err != nil {
		return nil, fmt.Errorf("generate signed request failed: %w", err)
	}

	req = req.WithContext(ctx)
	resp, err := cli.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	reply := raw.Reply()
	if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	if err = reply.GetError(); err != nil {
		return nil, err
	}
	return &PayInReply{data: reply.GetData()}, nil
}

func (cli *Client) QueryPayIn(ctx context.Context, payload *GetPayInRecordPayload) (*PayInRecord, error) {
	raw := payload.toRaw(cli.conf)
	req, err := raw.GenerateSignedRequest(cli.conf.Secret, cli.conf.Key)
	if err != nil {
		return nil, fmt.Errorf("generate signed request failed: %w", err)
	}

	req = req.WithContext(ctx)
	resp, err := cli.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	reply := raw.Reply()
	if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	if err = reply.GetError(); err != nil {
		return nil, err
	}
	return reply.GetData(), nil
}

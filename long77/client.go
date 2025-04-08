package long77

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type transport struct {
	inner   http.RoundTripper
	baseURL *url.URL
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	uri := t.baseURL.ResolveReference(req.URL)
	req.URL = uri
	return t.inner.RoundTrip(req)
}

type Client struct {
	http   *http.Client
	config *Config
}

type Config struct {
	BaseURL string // long77 base url

	NotifyURL string // Callback url
	ReturnURL string // Success url

	PartnerID string //
	Secret    string //
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
		config: &conf,
	}, nil
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

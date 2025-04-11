package ragapay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	httptransport "github.com/decode-ex/payment-sdk/internal/http_transport"
)

const (
	_BASE_URL = "https://checkout.ragapay.com"
)

type Client struct {
	http *http.Client

	conf *Config
}

type Config struct {
	SuccessURL string

	PublicID string
	Password string
}

func NewClient(conf Config) (*Client, error) {
	transport, err := httptransport.NewTransport(_BASE_URL)
	if err != nil {
		return nil, err
	}

	return &Client{
		http: &http.Client{
			Transport: transport,
		},
		conf: &conf,
	}, nil
}

func (cli *Client) newCheckoutSession(ctx context.Context, payload *rawCheckoutPayload) (*CheckoutResponse, error) {
	req, err := payload.GenerateSignedRequest(cli.conf)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	resp, err := cli.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errBody := CheckoutResponseError{}
		err := json.NewDecoder(resp.Body).Decode(&errBody)
		if err != nil || len(errBody.Errors) == 0 {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, errBody.Errors[0].ErrorMessage)
	}

	dec := json.NewDecoder(resp.Body)
	var respBody CheckoutResponse
	if err := dec.Decode(&respBody); err != nil {
		return nil, err
	}
	return &respBody, nil
}

func (cli *Client) MakePurchase(ctx context.Context, order *CheckoutOrder) (*CheckoutResponse, error) {
	if err := order.Validate(); err != nil {
		return nil, err
	}
	purchase := order.toRaw(cli.conf)
	return cli.newCheckoutSession(ctx, purchase)
}

func (cli *Client) MakeDebit(ctx context.Context, order *CheckoutOrder) (string, error) {
	return "", fmt.Errorf("unsupported")
}

func (cli *Client) MakeTransfer(ctx context.Context, order *CheckoutOrder) (string, error) {
	return "", fmt.Errorf("unsupported")
}

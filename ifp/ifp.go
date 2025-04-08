package ifp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	http   *http.Client
	config *Config
}

type Config struct {
	BaseURL    string
	AccessKey  string
	PrivateKey []byte

	CallbackURL string
}

type transport struct {
	inner   http.RoundTripper
	baseURL *url.URL
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	uri := t.baseURL.ResolveReference(req.URL)
	req.URL = uri
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
		config: &conf,
	}, nil
}

// 买入指定金额
func (cli *Client) BuyWithAmount(ctx context.Context, req *FiatBuyRequest) (*BuyCoinReply, error) {
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
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	res := &BuyResponse{}
	err = json.NewDecoder(resp.Body).Decode(res)
	if err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return &BuyCoinReply{
		RedirectURL:       res.Data.RedirectURL,
		AdvertisementCode: res.Data.AdvertisementCode,
	}, nil
}

// 买入指定数量
func (cli *Client) BuyWithQuantity(ctx context.Context, req any) (*BuyCoinReply, error) {
	panic("not implemented")
}

func (cli *Client) QueryOrder(ctx context.Context, req *QueryOrderRequest) (*QueryOrderResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	reqBody, err := req.toRaw(cli.config).GenerateSignedRequest(ctx, cli.config)
	if err != nil {
		return nil, err
	}

	res, err := cli.http.Do(reqBody)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	body := &QueryOrderResponse{}
	err = json.NewDecoder(res.Body).Decode(body)
	if err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}
	return body, nil
}

type baseRequest struct {
	ts string // timestamp, millisecond
}

func (base *baseRequest) GenerateSignature(accessKey string, privateKey []byte) string {
	content := fmt.Sprintf("%s_%s", accessKey, base.ts)
	mac := hmac.New(sha256.New, privateKey)
	mac.Write([]byte(content))

	sign := hex.EncodeToString(mac.Sum(nil))

	return strings.ToUpper(sign)
}

func newBaseRequest() *baseRequest {
	return &baseRequest{
		ts: strconv.FormatInt(time.Now().UnixMilli(), 10),
	}
}

func newBaseRequestWithTimestamp(ts int64) *baseRequest {
	return &baseRequest{
		ts: strconv.FormatInt(ts, 10),
	}
}

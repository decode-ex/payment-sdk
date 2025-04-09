package httptransport

import (
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

func NewTransport(baseURL string) (http.RoundTripper, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &transport{
		inner:   http.DefaultTransport,
		baseURL: parsedURL,
	}, nil
}

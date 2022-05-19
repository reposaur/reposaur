package client

import (
	"io"
	"net/http"
	"time"
)

const retryAfterHeader = "Retry-After"

var DefaultClient = &defaultClient{httpClient: http.DefaultClient}

type Client interface {
	Client() *http.Client
	Do(*http.Request) (*http.Response, error)
	NewRequest(method, path string, body io.Reader) (*http.Request, error)
}

type ThrottleTransport struct {
	transport http.RoundTripper
}

func NewThrottleTransport(transport http.RoundTripper) *ThrottleTransport {
	if transport == nil {
		transport = http.DefaultTransport
	}

	return &ThrottleTransport{
		transport: transport,
	}
}

func (t ThrottleTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	retryAfterVal := resp.Header.Get(retryAfterHeader)
	if retryAfterVal == "" {
		return resp, nil
	}

	retryAfter, err := time.ParseDuration(retryAfterVal + "s")
	if err != nil {
		return nil, err
	}

	time.Sleep(retryAfter)

	return t.RoundTrip(req)
}

type defaultClient struct {
	httpClient *http.Client
}

func (c defaultClient) Client() *http.Client {
	return c.httpClient
}

func (c defaultClient) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func (c defaultClient) NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, path, body)
}

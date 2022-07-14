package client

import (
	"context"
	"net/http"
	"net/url"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"
)

const (
	DefaultBaseURL = "https://api.github.com"
)

type Client struct {
	BaseURL *url.URL

	client       *retryablehttp.Client
	appTransport *ghinstallation.Transport
	token        string
}

func NewClient(httpClient *http.Client) *Client {
	baseURL, _ := url.Parse(DefaultBaseURL)
	client := newRetryableClient(httpClient)

	return &Client{
		BaseURL: baseURL,
		client:  client,
	}
}

func NewTokenClient(ctx context.Context, token string) *Client {
	tokenSrc := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})

	oauthClient := oauth2.NewClient(ctx, tokenSrc)

	client := NewClient(oauthClient)
	client.token = token

	return client
}

func NewAppClient(ctx context.Context, baseURL string, appID, installationID int64, privateKey []byte) (*Client, error) {
	appTransport, err := ghinstallation.New(nil, appID, installationID, privateKey)
	if err != nil {
		return nil, err
	}

	appTransport.BaseURL = baseURL

	httpClient := &http.Client{
		Transport: appTransport,
	}

	client := NewClient(httpClient)
	client.appTransport = appTransport

	return client, nil
}

func (c Client) Client() *http.Client {
	return c.client.HTTPClient
}

func (c Client) Token(ctx context.Context) (string, error) {
	if c.appTransport != nil {
		return c.appTransport.Token(ctx)
	}

	return c.token, nil
}

func (c Client) NewRequest(method, path string, rawBody any) (*retryablehttp.Request, error) {
	u, err := c.BaseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	req, err := retryablehttp.NewRequest(method, u.String(), rawBody)
	if err != nil {
		return nil, err
	}

	if rawBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("User-Agent", "reposaur")

	return req, nil
}

func (c Client) Do(req *retryablehttp.Request) (*http.Response, error) {
	return c.client.Do(req)
}

func newRetryableClient(httpClient *http.Client) *retryablehttp.Client {
	client := retryablehttp.NewClient()

	if httpClient != nil {
		client.HTTPClient = httpClient
	}

	return client
}

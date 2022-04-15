package github

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/oauth2"
)

const GITHUB_HOST = "api.github.com"

type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		token := getEnvVar("", "GITHUB_TOKEN", "GH_TOKEN")

		if token == "" {
			httpClient = http.DefaultClient
		} else {
			ts := oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: token,
				},
			)

			httpClient = oauth2.NewClient(context.Background(), ts)
		}
	}

	host := getEnvVar(GITHUB_HOST, "GITHUB_HOST", "GH_HOST")
	baseURL, _ := url.Parse("https://" + host)

	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (c Client) NewRequest(method string, urlStr string, body interface{}) (*http.Request, error) {
	url, err := c.baseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)

		enc.SetEscapeHTML(false)

		if err := enc.Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "reposaur")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func getEnvVar(def string, names ...string) string {
	for _, n := range names {
		val := os.Getenv(n)

		if val != "" {
			return val
		}
	}

	return def
}

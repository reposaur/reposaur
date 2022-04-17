package util

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"golang.org/x/oauth2"
)

const defaultGitHubHost = "api.github.com"

type githubTransporter struct {
	tr http.RoundTripper
}

func (t githubTransporter) RoundTrip(req *http.Request) (*http.Response, error) {
	ghHost := defaultGitHubHost

	if host := GetEnv("GITHUB_HOST", "GH_HOST"); host != nil {
		ghHost = *host
	}

	req.URL.Host = ghHost
	req.URL.Scheme = "https"

	return t.tr.RoundTrip(req)
}

// NewTokenHTTPClient creates an http.Client with a
// oauth2.StaticTokenSource using the provided token.
func NewTokenHTTPClient(ctx context.Context, token string) *http.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	)

	tc := oauth2.NewClient(ctx, ts)

	httpClient := &http.Client{
		Transport: githubTransporter{
			tr: tc.Transport,
		},
	}

	return httpClient
}

// NewInstallationHTTPClient creates an http.Client with authenticated
// using an app's installation token. The token is refreshed
// automatically.
//
// The Private Key provided must be Base64 encoded.
func NewInstallationHTTPClient(ctx context.Context, appID, installationID int64, appPrivKey string) (*http.Client, error) {
	// private key is base64 encoded
	privKey, err := base64.RawStdEncoding.DecodeString(appPrivKey)
	if err != nil {
		return nil, err
	}

	itr, err := ghinstallation.New(http.DefaultTransport, appID, installationID, privKey)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Transport: githubTransporter{
			tr: itr,
		},
	}

	return httpClient, nil
}

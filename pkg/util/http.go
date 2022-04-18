package util

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/gregjones/httpcache"
	"golang.org/x/oauth2"
)

const defaultGitHubHost = "api.github.com"

type githubTransport struct {
	Transport http.RoundTripper
}

func (t githubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ghHost := defaultGitHubHost

	if host := GetEnv("GITHUB_HOST", "GH_HOST"); host != nil {
		ghHost = *host
	}

	req.URL.Host = ghHost
	req.URL.Scheme = "https"

	return t.Transport.RoundTrip(req)
}

// NewTokenHTTPClient creates an http.Client with a
// oauth2.StaticTokenSource using the provided token.
func NewTokenHTTPClient(ctx context.Context, token string) *http.Client {
	ghTransport := githubTransport{
		Transport: http.DefaultTransport,
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Transport: ghTransport,
	})

	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	)

	tokenTransport := oauth2.NewClient(ctx, tokenSource).Transport

	cacheTransport := httpcache.NewMemoryCacheTransport()
	cacheTransport.Transport = tokenTransport

	return cacheTransport.Client()
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

	ghTransport := githubTransport{
		Transport: http.DefaultTransport,
	}

	installationTransport, err := ghinstallation.New(ghTransport, appID, installationID, privKey)
	if err != nil {
		return nil, err
	}

	cacheTransport := httpcache.NewMemoryCacheTransport()
	cacheTransport.Transport = installationTransport

	return cacheTransport.Client(), nil
}

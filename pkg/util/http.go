package util

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/gregjones/httpcache"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

const (
	defaultGitHubHost = "api.github.com"
	retryAfterHeader  = "Retry-After"
)

type githubTransport struct {
	logger    zerolog.Logger
	transport http.RoundTripper
}

func (t githubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ghHost := defaultGitHubHost

	if host := GetEnv("GITHUB_HOST", "GH_HOST"); host != nil {
		ghHost = *host
	}

	req.URL.Host = ghHost
	req.URL.Scheme = "https"

	t.logger.Debug().
		Str("method", req.Method).
		Str("url", req.URL.String()).
		Msg("Sending request to GitHub")

	return t.throttle(req)
}

func (t *githubTransport) throttle(req *http.Request) (*http.Response, error) {
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

	logger := t.logger.With().
		Str("path", req.URL.Path).
		Dur("retry after", retryAfter).
		Logger()

	logger.Info().Msg("Hit secondary rate limit. Waiting before trying...")
	time.Sleep(retryAfter)
	logger.Info().Msg("Continuing...")

	return t.RoundTrip(req)
}

// NewTokenHTTPClient creates an http.Client with a
// oauth2.StaticTokenSource using the provided token.
func NewTokenHTTPClient(ctx context.Context, logger zerolog.Logger, token string) *http.Client {
	ghTransport := &githubTransport{
		logger:    logger,
		transport: http.DefaultTransport,
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
func NewInstallationHTTPClient(ctx context.Context, logger zerolog.Logger, appID, installationID int64, appPrivKey string) (*http.Client, error) {
	// private key is base64 encoded
	privKey, err := base64.RawStdEncoding.DecodeString(appPrivKey)
	if err != nil {
		return nil, err
	}

	ghTransport := githubTransport{
		logger:    logger,
		transport: http.DefaultTransport,
	}

	installationTransport, err := ghinstallation.New(ghTransport, appID, installationID, privKey)
	if err != nil {
		return nil, err
	}

	cacheTransport := httpcache.NewMemoryCacheTransport()
	cacheTransport.Transport = installationTransport

	return cacheTransport.Client(), nil
}

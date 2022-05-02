package sdk

import (
	"context"
	"net/http"
	"os"

	"github.com/reposaur/reposaur/internal/builtins"
	"github.com/reposaur/reposaur/internal/policy"
	"github.com/reposaur/reposaur/pkg/output"
	"github.com/reposaur/reposaur/pkg/util"
	"github.com/rs/zerolog"
)

// Option represents a Reposaur option that can change a
// particular behavior.
type Option func(*Reposaur)

// Reposaur represents an instance of the auditing engine. It can be
// started with several options that control configuration, logging and
// the client to GitHub.
type Reposaur struct {
	logger     zerolog.Logger
	engine     *policy.Engine
	httpClient *http.Client
}

// New returns a new Reposaur instance, loading and
// compiling any policies provided and registering
// the built-in functions.
//
// If an HTTP client isn't passed as an option, a default
// client is created. A default (unauthenticated) client is created
// using `util.GitHubTransport`.
//
// The util functions available in the `util` package can be used to
// create authenticated HTTP clients.
//
// The default HTTP client will use the default host `api.github.com`. Can
// be customized using the `GITHUB_HOST` or `GH_HOST` environment variables.
func New(ctx context.Context, policyPaths []string, opts ...Option) (*Reposaur, error) {
	sdk := &Reposaur{
		logger: zerolog.New(os.Stderr),
	}

	for _, opt := range opts {
		opt(sdk)
	}

	if sdk.httpClient == nil {
		sdk.httpClient = &http.Client{
			Transport: util.GitHubTransport{
				Logger:    sdk.logger,
				Transport: http.DefaultTransport,
			},
		}
	}

	builtins.RegisterBuiltins(sdk.httpClient)

	var err error

	if len(policyPaths) == 0 {
		sdk.engine = policy.New()
	} else {
		sdk.engine, err = policy.Load(ctx, policyPaths)
		if err != nil {
			return nil, err
		}
	}

	return sdk, nil
}

// WithLogger sets the logger used by Reposaur.
func WithLogger(logger zerolog.Logger) Option {
	return func(sdk *Reposaur) {
		sdk.logger = logger
	}
}

// WithHTTPClient sets the HTTP client used by Reposaur's
// built-in functions.
func WithHTTPClient(client *http.Client) Option {
	return func(sdk *Reposaur) {
		sdk.httpClient = client
	}
}

// Logger returns Reposaur's logger.
func (sdk Reposaur) Logger() zerolog.Logger {
	return sdk.logger
}

// Client returns Reposaur's GitHub client.
func (sdk Reposaur) HTTPClient() *http.Client {
	return sdk.httpClient
}

// Engine returns Reposaur's policy engine.
func (sdk Reposaur) Engine() *policy.Engine {
	return sdk.engine
}

// Check executes the policies loaded with namespace against data
func (sdk Reposaur) Check(ctx context.Context, namespace string, data interface{}) (output.Report, error) {
	report, err := sdk.engine.Check(ctx, namespace, data)
	if err != nil {
		return output.Report{}, err
	}

	return report, nil
}

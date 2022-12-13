package sdk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/open-policy-agent/opa/compile"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/tester"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/reposaur/reposaur/internal/policy"
	"github.com/reposaur/reposaur/pkg/output"
	"github.com/reposaur/reposaur/provider"
	"github.com/reposaur/reposaur/provider/github"

	"github.com/rs/zerolog"
)

var DefaultProviders = []provider.Provider{
	github.NewProvider(nil),
}

// Option represents a Reposaur option that can change a
// particular behavior.
type Option func(*Reposaur)

// Reposaur represents an instance of the auditing engine. It can be
// started with several options that control configuration, logging and
// the client to GitHub.
type Reposaur struct {
	logger        zerolog.Logger
	engine        *policy.Engine
	providers     []provider.Provider
	enableTracing bool
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

	if len(sdk.providers) == 0 {
		sdk.providers = DefaultProviders
	}

	for _, p := range sdk.providers {
		for _, b := range p.Builtins() {
			rego.RegisterBuiltinDyn(b.Func(), b.Impl)
		}
	}

	var err error
	sdk.engine, err = policy.Load(ctx, policyPaths, policy.WithTracingEnabled(sdk.enableTracing))
	if err != nil {
		return nil, err
	}

	return sdk, nil
}

// WithProvider adds a provider to Reposaur.
func WithProvider(provider provider.Provider) Option {
	return func(sdk *Reposaur) {
		sdk.providers = append(sdk.providers, provider)
	}
}

// WithLogger sets the logger used by Reposaur.
func WithLogger(logger zerolog.Logger) Option {
	return func(sdk *Reposaur) {
		sdk.logger = logger
	}
}

// WithTracingEnabled enables or disables policy
// execution tracing.
func WithTracingEnabled(enabled bool) Option {
	return func(sdk *Reposaur) {
		sdk.enableTracing = enabled
	}
}

// Logger returns Reposaur logger.
func (sdk Reposaur) Logger() zerolog.Logger {
	return sdk.logger
}

// Engine returns Reposaur policy engine.
func (sdk Reposaur) Engine() *policy.Engine {
	return sdk.engine
}

// Check executes the policies loaded against data. Data is checked against every
// provider to derive a namespace and additional report properties.
func (sdk Reposaur) Check(ctx context.Context, data interface{}) (output.Report, error) {
	var (
		dataProvider provider.Provider
		namespace    provider.Namespace
		err          error
	)

	for _, p := range sdk.providers {
		namespace, err = provider.DeriveNamespace(p, data)
		if err != nil {
			if errors.Is(err, provider.ErrNonDerivable) {
				continue
			}

			return output.Report{}, err
		}

		dataProvider = p
	}

	if dataProvider == nil {
		return output.Report{}, errors.New("could not derive a valid namespace from data")
	}

	report, err := sdk.engine.Check(ctx, string(namespace), data)
	if err != nil {
		return output.Report{}, err
	}

	report.Properties, err = provider.DeriveProperties(dataProvider, namespace, data)
	if err != nil && !errors.Is(err, provider.ErrNonDerivable) {
		return output.Report{}, err
	}

	return report, nil
}

func (sdk Reposaur) Test(ctx context.Context) ([]*tester.Result, error) {
	runner := tester.NewRunner().
		EnableTracing(sdk.enableTracing).
		CapturePrintOutput(true).
		SetCompiler(sdk.engine.Compiler()).
		SetModules(sdk.engine.Modules())

	ch, err := runner.RunTests(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("running tests: %w", err)
	}

	var rawResults []*tester.Result
	for result := range ch {
		if result.Error != nil {
			return nil, fmt.Errorf("run test: %w", result.Error)
		}

		rawResults = append(rawResults, result)
		buf := new(bytes.Buffer)
		topdown.PrettyTrace(buf, result.Trace)

		var traces []string
		for _, line := range strings.Split(buf.String(), "\n") {
			if len(line) > 0 {
				traces = append(traces, line)
			}
		}

		for _, t := range traces {
			fmt.Fprintln(os.Stderr, t)
		}
	}

	return rawResults, nil
}

// Bundle builds a new OCI-compatible policy bundle.
func (sdk Reposaur) Bundle(ctx context.Context, paths []string, out io.Writer) error {
	c := compile.New().
		WithOutput(out).
		WithTarget("rego").
		WithPaths(paths...)

	if err := c.Build(ctx); err != nil {
		return err
	}

	return nil
}

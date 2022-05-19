package sdk

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/open-policy-agent/opa/tester"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/reposaur/reposaur/internal/policy"
	"github.com/reposaur/reposaur/pkg/output"
	"github.com/rs/zerolog"
)

// Option represents a Reposaur option that can change a
// particular behavior.
type Option func(*Reposaur)

// Reposaur represents an instance of the auditing engine. It can be
// started with several options that control configuration, logging and
// the client to GitHub.
type Reposaur struct {
	logger        zerolog.Logger
	engine        *policy.Engine
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

	var err error
	sdk.engine, err = policy.Load(ctx, policyPaths, policy.WithTracingEnabled(sdk.enableTracing))
	if err != nil {
		return nil, err
	}

	return sdk, nil
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

// Logger returns Reposaur's logger.
func (sdk Reposaur) Logger() zerolog.Logger {
	return sdk.logger
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

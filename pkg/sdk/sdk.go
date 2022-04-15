package sdk

import (
	"context"

	"github.com/reposaur/reposaur/internal/builtins"
	"github.com/reposaur/reposaur/internal/policy"
	"github.com/reposaur/reposaur/pkg/github"
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
	logger zerolog.Logger
	client *github.Client
	engine *policy.Engine
}

// New returns a new Reposaur instance.
func New(ctx context.Context, policyPaths []string, opts ...Option) (*Reposaur, error) {
	sdk := &Reposaur{
		logger: zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger(),
	}

	for _, opt := range opts {
		opt(sdk)
	}

	if sdk.client == nil {
		sdk.client = github.NewClient(nil)
	}

	builtins.RegisterBuiltins(sdk.client)

	var err error

	sdk.engine, err = policy.Load(ctx, policyPaths)
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

// WithGitHubClient sets the GitHub client used by Reposaur.
func WithGitHubClient(client *github.Client) Option {
	return func(sdk *Reposaur) {
		sdk.client = client
	}
}

// Logger returns Reposaur's logger.
func (sdk Reposaur) Logger() zerolog.Logger {
	return sdk.logger
}

// Client returns Reposaur's GitHub client.
func (sdk Reposaur) Client() *github.Client {
	return sdk.client
}

// Engine returns Reposaur's policy engine.
func (sdk Reposaur) Engine() *policy.Engine {
	return sdk.engine
}

// Check executes the policies loaded against one or more
// fetchable GitHub objects.
func (sdk Reposaur) Check(ctx context.Context, data map[string]interface{}) (output.Report, error) {
	var reports []output.Report

	for n, d := range data {
		r, err := sdk.engine.Check(ctx, n, d)
		if err != nil {
			return output.Report{}, err
		}

		reports = append(reports, r)
	}

	return mergeReports(reports), nil
}

func mergeReports(reports []output.Report) output.Report {
	report := output.Report{
		Rules:   make(map[string]*output.Rule),
		Results: make(map[string]*output.Result),
	}

	for _, r := range reports {
		report.RuleCount += r.RuleCount

		for k, v := range r.Rules {
			report.Rules[k] = v
		}

		for k, v := range r.Results {
			report.Results[k] = v
		}
	}

	return report
}

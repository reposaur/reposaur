package test

import (
	"context"
	"os"
	"time"

	"github.com/reposaur/reposaur/cmd/rsr/internal/cmdutil"
	"github.com/reposaur/reposaur/pkg/sdk"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type testParams struct {
	policyPaths    []string
	outputFilename string
	enableTracing  bool
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test POLICY_PATH...",
		Short: "Runs the tests available in POLICY_PATH",
		Long:  "Runs the tests available in POLICY_PATH",
	}

	var (
		params = &testParams{policyPaths: []string{"."}}
		flags  = cmd.Flags()
	)

	cmdutil.AddOutputFlag(flags, &params.outputFilename)
	cmdutil.AddTraceFlag(flags, &params.enableTracing)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		var (
			ctx    = cmd.Context()
			logger = zerolog.Ctx(ctx)
		)

		opts := []sdk.Option{
			sdk.WithLogger(*logger),
			sdk.WithTracingEnabled(params.enableTracing),
		}

		if len(args) > 0 {
			params.policyPaths = args
		}

		rsr, err := sdk.New(ctx, params.policyPaths, opts...)
		if err != nil {
			logger.Fatal().Err(err).Msg("could not instantiate SDK")
		}

		runTest(ctx, rsr)
	}

	return cmd
}

// runTest executes policy tests, outputting the results
// to the provided  outWriter.
//
// If any test fails, the function will exit with code 1.
// Otherwise, exits with code 0.
func runTest(ctx context.Context, rsr *sdk.Reposaur) {
	var (
		startTime = time.Now()
		logger    = zerolog.Ctx(ctx)
	)

	results, err := rsr.Test(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to execute tests")
	}

	var failedTests, totalTests int

	for _, r := range results {
		totalTests++

		if r.Fail {
			failedTests++
			logger.Error().Msg(r.String())
		} else {
			logger.Info().Msg(r.String())
		}
	}

	testLogger := logger.With().
		Int("passed", totalTests-failedTests).
		Int("failed", failedTests).
		Int("total", totalTests).
		Dur("timeElapsed", time.Since(startTime)).
		Logger()

	if failedTests > 0 {
		testLogger.Error().Msg("done")
		os.Exit(1)
	}

	testLogger.Info().Msg("done")
	os.Exit(0)
}

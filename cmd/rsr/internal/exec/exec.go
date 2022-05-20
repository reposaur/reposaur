package exec

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/reposaur/reposaur/cmd/rsr/internal/cmdutil"
	"github.com/reposaur/reposaur/pkg/output"
	"github.com/reposaur/reposaur/pkg/sdk"
	"github.com/reposaur/reposaur/provider/github"
	githubclient "github.com/reposaur/reposaur/provider/github/client"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type execParams struct {
	policyPaths    []string
	outputFilename string
	inputFilename  string
	enableTracing  bool
	github         cmdutil.GitHubClientOptions
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec [-p POLICY_PATH...] [-n NAMESPACE] [-o OUTPUT] INPUT",
		Short: "Executes policies against INPUT data",
		Long:  "Executes policies against INPUT data",
	}

	var (
		params = &execParams{}
		flags  = cmd.Flags()
	)

	cmdutil.AddOutputFlag(flags, &params.outputFilename)
	cmdutil.AddPolicyPathsFlag(flags, &params.policyPaths)
	cmdutil.AddTraceFlag(flags, &params.enableTracing)
	cmdutil.AddGitHubFlags(flags, &params.github)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		var (
			ctx    = cmd.Context()
			logger = zerolog.Ctx(ctx)
		)

		if len(args) > 1 {
			logger.Fatal().Strs("args", args).Msg("too many arguments for INPUT")
		}

		if len(args) == 1 {
			params.inputFilename = args[0]
		}

		inReader, err := cmdutil.GetInputReader(ctx, params.inputFilename)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to get input reader")
		}
		defer inReader.Close()

		outWriter, err := cmdutil.GetOutputWriter(ctx, params.outputFilename)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to get output writer")
		}
		defer outWriter.Close()

		githubProvider, err := newGitHubProvider(ctx, &params.github)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to create GitHub provider")
		}

		opts := []sdk.Option{
			sdk.WithLogger(*logger),
			sdk.WithProvider(githubProvider),
			sdk.WithTracingEnabled(params.enableTracing),
		}

		rsr, err := sdk.New(ctx, params.policyPaths, opts...)
		if err != nil {
			logger.Fatal().Err(err).Msg("could not instantiate SDK")
		}

		runExec(ctx, rsr, inReader, outWriter)
	}

	return cmd
}

// runExec will execute the policies against the data available
// in inReader. The resulting reports will be outputted to outWriter.
func runExec(ctx context.Context, rsr *sdk.Reposaur, inReader io.ReadCloser, outWriter io.WriteCloser) {
	startTime := time.Now()

	var (
		inputsCh = make(chan interface{})
		inputsWg = sync.WaitGroup{}

		reportsCh = make(chan output.Report)
		reportsWg = sync.WaitGroup{}

		logger = zerolog.Ctx(ctx)
	)

	// Process inputs
	go func() {
		for input := range inputsCh {
			inputsWg.Done()
			reportsWg.Add(1)

			go func(input interface{}) {
				logger.Debug().Msg("processing input")

				report, err := rsr.Check(ctx, input)
				if err != nil {
					logger.Fatal().Err(err).Send()
				}

				logger.Debug().Msg("done processing input")

				reportsCh <- report
			}(input)
		}
	}()

	// Output reports
	go func() {
		enc := json.NewEncoder(outWriter)
		enc.SetIndent("", "  ")

		for report := range reportsCh {
			sarif, err := output.NewSarifReport(report)
			if err != nil {
				logger.Fatal().Err(err).Send()
			}

			if err := enc.Encode(sarif); err != nil {
				logger.Fatal().Err(err).Send()
			}

			reportsWg.Done()
		}
	}()

	// Start processing inputs
	dec := json.NewDecoder(inReader)

	for {
		var input interface{}

		if err := dec.Decode(&input); err == io.EOF {
			break
		} else if err != nil {
			logger.Fatal().Err(err).Msg("failed to decode input")
		}

		switch inputT := input.(type) {
		case map[string]interface{}:
			logger.Debug().Msg("received 1 input")

			inputsWg.Add(1)
			inputsCh <- inputT

		case []interface{}:
			logger.Debug().Msgf("received %d inputs", len(inputT))

			for _, input := range inputT {
				inputsWg.Add(1)
				inputsCh <- input
			}
		}
	}

	inputsWg.Wait()
	close(inputsCh)
	logger.Debug().Msg("closed inputs channel")

	reportsWg.Wait()
	close(reportsCh)
	logger.Debug().Msg("closed reports channel")

	logger.Info().Dur("timeElappsed", time.Since(startTime)).Msg("done")

	// TODO: should exit with 1 if there are failed results
	os.Exit(0)
}

func newGitHubProvider(ctx context.Context, opts *cmdutil.GitHubClientOptions) (*github.GitHub, error) {
	var client *githubclient.Client

	if opts.AppID != 0 && opts.InstallationID != 0 && opts.AppPrivateKey != "" {
		appClient, err := githubclient.NewAppClient(ctx, opts.BaseURL, opts.AppID, opts.InstallationID, []byte(opts.AppPrivateKey))
		if err != nil {
			return nil, err
		}

		client = appClient
	} else if opts.Token != "" {
		client = githubclient.NewTokenClient(ctx, opts.Token)
	}

	return github.NewProvider(client), nil
}

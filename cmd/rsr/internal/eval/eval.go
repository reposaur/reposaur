package eval

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"
	"github.com/reposaur/reposaur/reposaur"
	"golang.org/x/exp/slog"
	"io"
	"os"
	"sync"
	"time"

	"github.com/reposaur/reposaur/cmd/rsr/internal/cmdutil"
	"github.com/reposaur/reposaur/provider/github"
	githubclient "github.com/reposaur/reposaur/provider/github/client"
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
		Use:   "eval [-p POLICY...] [-o OUTPUT] INPUT",
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
			logger = slog.FromContext(ctx)
		)

		if len(args) > 1 {
			logger.Error("too many arguments for INPUT", nil, "input", args)
			os.Exit(1)
		}

		if len(args) == 1 {
			params.inputFilename = args[0]
		}

		inReader, err := cmdutil.GetInputReader(ctx, params.inputFilename)
		if err != nil {
			logger.Error("failed to open input reader", err)
			os.Exit(1)
		}
		defer inReader.Close()

		outWriter, err := cmdutil.GetOutputWriter(ctx, params.outputFilename)
		if err != nil {
			logger.Error("failed to open output writer", err)
			os.Exit(1)
		}
		defer outWriter.Close()

		githubProvider, err := newGitHubProvider(ctx, &params.github)
		if err != nil {
			logger.Error("failed to create GitHub provider", err)
			os.Exit(1)
		}

		rsr, err := reposaur.New(reposaur.WithProviders(githubProvider))
		if err != nil {
			logger.Error("failed to instantiate Reposaur", err)
			os.Exit(1)
		}

		if err := rsr.LoadPolicies(params.policyPaths...); err != nil {
			if errors.As(err, &loader.Errors{}) || errors.As(err, &ast.Errors{}) {
				_, _ = fmt.Fprintln(os.Stderr, err.Error())
				err = nil
			}
			logger.Error("failed to load policies", err)
			os.Exit(1)
		}

		runEval(ctx, rsr, inReader, outWriter)
	}

	return cmd
}

// runEval will execute the policies against the data available
// in inReader. The resulting reports will be outputted to outWriter.
func runEval(ctx context.Context, rsr *reposaur.Reposaur, inReader io.ReadCloser, outWriter io.WriteCloser) {
	startTime := time.Now()

	var (
		inputsCh = make(chan any)
		inputsWg = sync.WaitGroup{}

		reportsCh = make(chan *reposaur.Report)
		reportsWg = sync.WaitGroup{}

		logger = slog.FromContext(ctx)
	)

	// Process inputs
	go func() {
		processInputs(ctx, rsr, inputsCh, &inputsWg, reportsCh, &reportsWg)
	}()

	// Output reports
	go func() {
		outputReports(ctx, outWriter, reportsCh, &reportsWg)
	}()

	if err := readInput(ctx, inReader, inputsCh, &inputsWg); err != nil {
		logger.Error("failed to decode input", err)
		os.Exit(1)
	}

	inputsWg.Wait()
	close(inputsCh)
	logger.Debug("closed inputs channel")

	reportsWg.Wait()
	close(reportsCh)
	logger.Debug("closed reports channel")

	logger.Info("done", "timeElapsed", time.Since(startTime))

	// TODO: should exit with 1 if there are failed results
	os.Exit(0)
}

func processInputs(ctx context.Context, rsr *reposaur.Reposaur, inputsCh chan any, inputsWg *sync.WaitGroup, reportsCh chan *reposaur.Report, reportsWg *sync.WaitGroup) {
	logger := slog.FromContext(ctx)

	for input := range inputsCh {
		inputsWg.Done()
		reportsWg.Add(1)

		go func(input any) {
			logger.Debug("processing input")

			report, err := rsr.Eval(ctx, input)
			if err != nil {
				logger.Error("failed to evaluate policies", err)
				reportsWg.Done()
				return
			}

			logger.Debug("done processing input")
			reportsCh <- report
		}(input)
	}
}

func outputReports(ctx context.Context, outWriter io.WriteCloser, reportsCh chan *reposaur.Report, reportsWg *sync.WaitGroup) {
	logger := slog.FromContext(ctx)
	enc := json.NewEncoder(outWriter)
	enc.SetIndent("", "  ")

	for report := range reportsCh {
		sarif, err := report.SARIF()
		if err != nil {
			logger.Error("failed to get SARIF report", err)
		}

		if err := enc.Encode(sarif); err != nil {
			logger.Error("failed to encode report as JSON", err)
		}

		reportsWg.Done()
	}
}

func readInput(ctx context.Context, inReader io.ReadCloser, inputsCh chan any, inputsWg *sync.WaitGroup) error {
	logger := slog.FromContext(ctx)

	for {
		var input any

		if err := json.NewDecoder(inReader).Decode(&input); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		switch inputT := input.(type) {
		case map[string]any:
			logger.Debug("received 1 input")
			inputsWg.Add(1)
			inputsCh <- inputT

		case []any:
			logger.Debug("received multiple inputs", "len", len(inputT))
			for _, input := range inputT {
				inputsWg.Add(1)
				inputsCh <- input
			}
		}
	}

	return nil
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

	return github.New(client), nil
}

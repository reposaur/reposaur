package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/reposaur/reposaur/internal/build"
	"github.com/reposaur/reposaur/pkg/detector"
	"github.com/reposaur/reposaur/pkg/output"
	"github.com/reposaur/reposaur/pkg/sdk"
	"github.com/reposaur/reposaur/pkg/util"
	"github.com/rs/zerolog"
)

const usage = `Usage:
    rsr [--exec] [-p POLICY_PATH] [-n NAMESPACE] [-o OUTPUT] [INPUT]
    rsr --test [-o OUTPUT]

Options:
    -e, --exec                 Executes the policies against the input data.
    -t, --test                 Runs the tests specified in the input.
    -T, --trace                Enables policy execution tracing.
    -o, --output OUTPUT        Write the output to the file at path OUTPUT.
    -p, --policy POLICY_PATH   Use the policies at POLICY_PATH. Can be repeated.
    -n, --namespace NAMESPACE  Custom policy namespace.
    -v, --verbose              Increases the logger verbosity.
    -V, --version              Prints version information.
    -h, --help                 Prints usage information.

INPUT defaults to standard input and OUTPUT defaults to standard output.
If OUTPUT exists, it will be overwritten.

POLICY_PATH can be a path to a single policy file or a directory of policy
files, defaults to current directory. Policies are loaded recursively.

NAMESPACE is the name of the policy package that will be executed. Namespaces
are detected automatically for repositories, organizations, users, issues and
pull requests.

Examples:
    # Execute policies against a static file
    $ cat repo.json | rsr

    # Fetch a single repository and execute policies
    $ gh api /repos/reposaur/reposaur | rsr

    # Fetch multiple repositories and execute policies
    $ gh api /orgs/reposaur/repos | rsr

    # Fetch all repositories and execute policies
    $ gh api /orgs/reposaur/repos --paginate | rsr

    # Using with other tools (i.e. to output total reports)
    $ gh api /orgs/reposaur/repos --paginate | rsr | jq -s length

    # Run policies tests
    $ rsr --test`

func main() {
	var (
		versionFlag, helpFlag, verboseFlag        bool
		execFlag, testFlag, traceFlag             bool
		outputFlag, policyPathFlag, namespaceFlag string
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "%s\n", usage)
	}

	// -V, --version
	flag.BoolVar(&versionFlag, "V", false, "")
	flag.BoolVar(&versionFlag, "version", false, "")

	// -h, --help
	flag.BoolVar(&helpFlag, "h", false, "")
	flag.BoolVar(&helpFlag, "help", false, "")

	// -e, --exec
	flag.BoolVar(&execFlag, "e", false, "")
	flag.BoolVar(&execFlag, "exec", false, "")

	// -t, --test
	flag.BoolVar(&testFlag, "t", false, "")
	flag.BoolVar(&testFlag, "test", false, "")

	// -T, --trace
	flag.BoolVar(&traceFlag, "T", false, "")
	flag.BoolVar(&traceFlag, "trace", false, "")

	// -o, --output
	flag.StringVar(&outputFlag, "o", "", "")
	flag.StringVar(&outputFlag, "output", "", "")

	// -p, --policy
	flag.StringVar(&policyPathFlag, "p", "./", "")
	flag.StringVar(&policyPathFlag, "policy", "./", "")

	// -n, --namespace
	flag.StringVar(&namespaceFlag, "n", "", "")
	flag.StringVar(&namespaceFlag, "namespace", "", "")

	// -v, --verbose
	flag.BoolVar(&verboseFlag, "v", false, "")
	flag.BoolVar(&verboseFlag, "verbose", false, "")

	flag.Parse()

	if helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	if versionFlag {
		fmt.Fprintf(os.Stdout, "version %s\n", build.Version)
		os.Exit(0)
	}

	var (
		logger = buildLogger(verboseFlag)
		ctx    = contextWithLogger(context.Background(), logger)
	)

	client, err := buildClient(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not create HTTP client")
	}

	opts := []sdk.Option{
		sdk.WithLogger(logger),
		sdk.WithHTTPClient(client),
		sdk.WithTracingEnabled(traceFlag),
	}

	rsr, err := sdk.New(ctx, []string{policyPathFlag}, opts...)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not instantiate SDK")
	}

	switch {
	case testFlag:
		if execFlag {
			logger.Fatal().Msg("-t/--test can't be used with -e/--exec")
		}

		if outputFlag != "" {
			logger.Fatal().Msg("-t/--test doesn't support custom -o/--output")
		}

		runTest(ctx, rsr)

	default:
		if testFlag {
			logger.Fatal().Msg("-e/--exec can't be used with -t/--test")
		}

		if flag.NArg() > 1 {
			logger.Fatal().Strs("args", flag.Args()).Msg("too many arguments for INPUT")
		}

		inReader, err := getInputReader(ctx, flag.Arg(0))
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to get input reader")
		}
		defer inReader.Close()

		outWriter, err := getOutputWriter(ctx, outputFlag)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to get output writer")
		}
		defer outWriter.Close()

		runExec(ctx, rsr, namespaceFlag, inReader, outWriter)
	}
}

// runExec will execute the policies against the data available
// in inReader. The resulting reports will be outputted to outWriter.
func runExec(ctx context.Context, rsr *sdk.Reposaur, namespace string, inReader io.ReadCloser, outWriter io.WriteCloser) {
	startTime := time.Now()

	var (
		inputsCh = make(chan interface{})
		inputsWg = sync.WaitGroup{}

		reportsCh = make(chan output.Report)
		reportsWg = sync.WaitGroup{}

		logger = loggerFromContext(ctx)
	)

	// Process inputs
	go func() {
		for input := range inputsCh {
			inputsWg.Done()
			reportsWg.Add(1)

			go func(input interface{}) {
				if namespace == "" {
					ns, err := detector.DetectNamespace(input)
					if err != nil {
						logger.Fatal().Err(err).Send()
					}
					namespace = ns
				}

				props, err := detector.DetectReportProperties(namespace, input)
				if err != nil {
					logger.Fatal().Err(err).Send()
				}

				processorLogger := logger.With().
					Interface("props", props).
					Str("namespace", namespace).
					Logger()

				processorLogger.Debug().Msg("processing input")

				report, err := rsr.Check(ctx, namespace, input)
				if err != nil {
					logger.Fatal().Err(err).Send()
				}
				report.Properties = props

				processorLogger.Debug().Msg("done processing input")

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

// runTest executes policy tests, outputting the results
// to the provided  outWriter.
//
// If any test fails, the function will exit with code 1.
// Otherwise exits with code 0.
func runTest(ctx context.Context, rsr *sdk.Reposaur) {
	var (
		startTime = time.Now()
		logger    = loggerFromContext(ctx)
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

	logger = logger.With().
		Int("passed", totalTests-failedTests).
		Int("failed", failedTests).
		Int("total", totalTests).
		Dur("timeEllapsed", time.Since(startTime)).
		Logger()

	if failedTests > 0 {
		logger.Error().Msg("done")
		os.Exit(1)
	}

	logger.Info().Msg("done")
	os.Exit(0)
}

// getInputReader returns a io.ReadCloser. If filename
// is not empty, the file is opened and returned. Otherwise
// returns a reader from standard input.
func getInputReader(ctx context.Context, filename string) (r io.ReadCloser, err error) {
	var (
		logger = loggerFromContext(ctx)
		file   = os.Stdin
	)

	if filename == "" {
		logger.Debug().Msg("using standard input as INPUT")
	} else {
		logger.Debug().Msgf("using %s as INPUT", filename)

		file, err = os.Open(filename)
		if err != nil {
			return nil, err
		}
	}

	return file, nil
}

// getOutputWriter returns a io.WriteCloser. If filename
// is not empty, the file is opened and returned. Otherwise
// returns a writer to standard output.
func getOutputWriter(ctx context.Context, filename string) (w io.WriteCloser, err error) {
	var (
		logger = loggerFromContext(ctx)
		file   = os.Stdout
	)

	if filename == "" {
		logger.Debug().Msg("using standard output as OUTPUT")
	} else {
		logger.Debug().Msgf("using %s as OUTPUT", filename)

		file, err = os.OpenFile(filename, os.O_WRONLY+os.O_CREATE+os.O_TRUNC, 0o666)
		if err != nil {
			return nil, err
		}
	}

	return file, nil
}

// buildLogger returns a new logger that outputs to
// the standard error output. If verbose is true,
// the log level will be debug, otherwise will be info.
func buildLogger(verbose bool) zerolog.Logger {
	var (
		lvl = zerolog.InfoLevel
		cw  = zerolog.ConsoleWriter{Out: os.Stderr}
	)

	if verbose {
		lvl = zerolog.DebugLevel
	}

	return zerolog.New(cw).Level(lvl).With().Timestamp().Logger()
}

// buildClient returns a http.Client authenticated to use
// to call the GitHub API. If no authentication information is
// found in environment variables returns a http.DefaultClient.
//
// If GITHUB_TOKEN or GH_TOKEN environment variables exist,
// the value is used as token.
//
// If all environment variables below exist, uses that information
// to authenticate as a GitHub App installation:
//
// * GITHUB_APP_ID or GH_APP_ID
// * GITHUB_APP_PRIVATE_KEY or GH_APP_PRIVATE_KEY
// * GITHUB_INSTALLATION_TOKEN or GH_INSTALLATION_TOKEN
//
// Note that the GitHub App private key must be base64-encoded.
func buildClient(ctx context.Context) (*http.Client, error) {
	logger := loggerFromContext(ctx)
	token := util.GetEnv("GITHUB_TOKEN", "GH_TOKEN")

	if token != nil {
		logger.Debug().Msg("found environment variable with GitHub token")
		return util.NewTokenHTTPClient(ctx, logger, *token), nil
	}

	var (
		appID = util.GetInt64Env("GITHUB_APP_ID", "GH_APP_ID")

		installationID = util.GetInt64Env(
			"GITHUB_INSTALLATION_TOKEN",
			"GH_INSTALLATION_TOKEN",
		)

		appPrivKey = util.GetEnv(
			"GITHUB_APP_PRIVATE_KEY",
			"GH_APP_PRIVATE_KEY",
		)
	)

	if appID != nil && installationID != nil && appPrivKey != nil {
		logger.Debug().Msg("found environment variables for GitHub App authentication")
		return util.NewInstallationHTTPClient(ctx, logger, *appID, *installationID, *appPrivKey)
	}

	logger.Debug().Msg("using an unauthenticated GitHub client")

	return http.DefaultClient, nil
}

// loggerKey is the key used to store a logger
// in a context.Context.
const loggerKey = "logger"

// contextWithLogger returns a new context from ctx with
// a logger assigned to the "logger" key.
func contextWithLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// loggerFromContext returns an existing logger from ctx.
// If ctx doesn't have a logger a new one is created using buildContext.
func loggerFromContext(ctx context.Context) zerolog.Logger {
	if logger, ok := ctx.Value(loggerKey).(zerolog.Logger); ok {
		return logger
	}

	return buildLogger(false)
}

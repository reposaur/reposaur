package root

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/reposaur/reposaur/pkg/cmd/exec"
	"github.com/reposaur/reposaur/pkg/cmd/test"
	"github.com/reposaur/reposaur/pkg/cmdutil"
	"github.com/reposaur/reposaur/pkg/util"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var cmd = &cobra.Command{
	Use:   "rsr",
	Short: "Executes a set of Rego policies against the data provided",
	Long:  "Executes a set of Rego policies against the data provided",
}

func NewCommand(f *cmdutil.Factory) *cobra.Command {
	execCmd := exec.NewCommand(f)

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		loggerLevel, err := cmd.Flags().GetString("logger-level")
		if err != nil {
			return err
		}

		loggerFormat, err := cmd.Flags().GetString("logger-format")
		if err != nil {
			return err
		}

		f.Logger, err = buildLogger(loggerLevel, loggerFormat)
		if err != nil {
			return err
		}

		f.HTTPClient, err = buildClient(cmd.Context(), f.Logger)
		if err != nil {
			return err
		}

		return nil
	}

	// Set the exec command as default
	cmd.RunE = execCmd.RunE

	execCmd.Flags().VisitAll(func(f *pflag.Flag) {
		cmd.Flags().AddFlag(f)
	})

	// Sub-commands
	cmd.AddCommand(execCmd)
	cmd.AddCommand(test.NewCommand(f))

	// Common flags
	cmd.PersistentFlags().String(
		"logger-format", "pretty",
		"logger format (one of 'pretty' and 'json')",
	)

	cmd.PersistentFlags().StringP(
		"logger-level", "l", "error",
		"logger level (one of 'info', 'warn', 'error' or 'debug')",
	)

	return cmd
}

func buildClient(ctx context.Context, logger zerolog.Logger) (*http.Client, error) {
	token := util.GetEnv(
		"GITHUB_TOKEN",
		"GH_TOKEN",
	)

	if token != nil {
		logger.Debug().Msg("Found environment variable with GitHub token")
		return util.NewTokenHTTPClient(ctx, logger, *token), nil
	}

	var (
		appID = util.GetInt64Env(
			"GITHUB_APP_ID",
			"GH_APP_ID",
		)

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
		logger.Debug().Msg("Found environment variables for GitHub App authentication")
		return util.NewInstallationHTTPClient(ctx, logger, *appID, *installationID, *appPrivKey)
	}

	logger.Debug().Msg("Using an unauthenticated GitHub client")

	return http.DefaultClient, nil
}

func buildLogger(level string, format string) (zerolog.Logger, error) {
	var logger zerolog.Logger

	switch level {
	case "info":
		logger = logger.Level(zerolog.InfoLevel)
		break
	case "warn":
		logger = logger.Level(zerolog.WarnLevel)
		break
	case "error":
		logger = logger.Level(zerolog.ErrorLevel)
		break
	case "debug":
		logger = logger.Level(zerolog.DebugLevel)
		break
	default:
		return logger, fmt.Errorf("unknown logger level '%s'", level)
	}

	switch format {
	case "pretty":
		cw := zerolog.NewConsoleWriter()
		cw.Out = os.Stderr
		logger = logger.Output(cw)

	case "json":
		logger = logger.Output(os.Stderr)

	default:
		return logger, fmt.Errorf("unknown logger format '%s'", format)
	}

	return logger, nil
}

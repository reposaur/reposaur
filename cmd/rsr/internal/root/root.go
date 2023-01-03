package root

import (
	"github.com/reposaur/reposaur/cmd/rsr/internal/cmdutil"
	"github.com/reposaur/reposaur/cmd/rsr/internal/eval"
	"github.com/reposaur/reposaur/internal/build"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
	"os"
)

type rootParams struct {
	verbose bool
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Version: build.Version,
		Use:     "rsr",
		Short:   "Reposaur - security & compliance for GitHub",
	}

	params := &rootParams{}

	cmdutil.AddVerboseFlag(cmd.PersistentFlags(), &params.verbose)

	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		loggerOpts := slog.HandlerOptions{
			Level: slog.LevelInfo,
		}

		if params.verbose {
			loggerOpts.Level = slog.LevelDebug
			loggerOpts.AddSource = true
		}

		logger := slog.New(
			loggerOpts.NewTextHandler(os.Stderr).WithGroup("reposaur"),
		)

		cmd.SetContext(slog.NewContext(cmd.Context(), logger))
	}

	cmd.AddCommand(
		eval.NewCmd(),
	)

	return cmd
}

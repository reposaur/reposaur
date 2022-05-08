package root

import (
	"context"

	"github.com/reposaur/reposaur/cmd/rsr/internal/cmdutil"
	"github.com/reposaur/reposaur/cmd/rsr/internal/exec"
	"github.com/reposaur/reposaur/cmd/rsr/internal/test"
	"github.com/reposaur/reposaur/internal/build"
	"github.com/spf13/cobra"
)

type rootParams struct {
	verbose bool
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Version: build.Version,
		Use:     "rsr",
		Short:   "Reposaur - security & compliance for GitHub metadata",
	}

	params := &rootParams{}

	cmdutil.AddVerboseFlag(cmd.PersistentFlags(), &params.verbose)

	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		cmd.SetContext(
			cmdutil.ContextWithLogger(
				context.Background(),
				cmdutil.NewLogger(params.verbose),
			),
		)
	}

	cmd.AddCommand(
		exec.NewCmd(),
		test.NewCmd(),
	)

	return cmd
}

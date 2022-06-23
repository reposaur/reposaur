package bundle

import (
	"os"

	"github.com/reposaur/reposaur/cmd/rsr/internal/cmdutil"
	"github.com/reposaur/reposaur/pkg/sdk"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type bundleParams struct {
	policyPaths    []string
	experimental bool
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bundle [-p POLICY_PATH...] <OUTPUT>",
		Short: "Creates a bundle from the policies at POLICY_PATH (experimental)",
		Long:  "Creates a bundle from the policies at POLICY_PATH (experimental)",
	}

	var (
		params = &bundleParams{}
		flags  = cmd.Flags()
	)

	cmdutil.AddPolicyPathsFlag(flags, &params.policyPaths)
	cmdutil.AddExperimentalFlag(flags, &params.experimental)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		var (
			ctx    = cmd.Context()
			logger = zerolog.Ctx(ctx)
		)

		if !params.experimental {
			logger.Fatal().Msg("experimental feature, please use --experimental to accept")
		}

		if len(args) != 1 {
			logger.Fatal().Msgf("exactly 1 arguments required, got %d", len(args))
		}

		rsr, err := sdk.New(ctx, params.policyPaths, sdk.WithLogger(*logger))
		if err != nil {
			logger.Fatal().Err(err).Msg("could not instantiate SDK")
		}

		out, err := os.OpenFile(args[0], os.O_WRONLY+os.O_CREATE+os.O_TRUNC, 0o666)
		if err != nil {
			logger.Fatal().Err(err).Msg("could not open file")
		}

		err = rsr.Bundle(ctx, params.policyPaths, out)
		if err != nil {
			logger.Fatal().Err(err).Msg("could not create bundle")
		}

		logger.Info().Msgf("bundle written to %s", args[0])
	}

	return cmd
}

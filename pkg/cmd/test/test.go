package test

import (
	"fmt"
	"os"

	"github.com/reposaur/reposaur/pkg/cmdutil"
	"github.com/reposaur/reposaur/pkg/sdk"
	"github.com/spf13/cobra"
)

type Params struct {
	policyPaths   []string
	enableTracing bool
}

var cmd = &cobra.Command{
	Use:   "test",
	Short: "Execute test cases",
	Long:  "Execute test cases",
}

func NewCommand(f *cmdutil.Factory) *cobra.Command {
	params := Params{}

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		rs, err := sdk.New(
			cmd.Context(),
			params.policyPaths,
			sdk.WithLogger(f.Logger),
			sdk.WithHTTPClient(f.HTTPClient),
			sdk.WithTracingEnabled(params.enableTracing),
		)
		if err != nil {
			return err
		}

		results, err := rs.Test(cmd.Context())
		if err != nil {
			return err
		}

		var fail bool

		for _, r := range results {
			if r.Fail {
				fail = true
			}

			fmt.Println(r.String())
		}

		if fail {
			os.Exit(1)
		}

		return nil
	}

	cmd.Flags().StringSliceVarP(
		&params.policyPaths,
		"policy", "p", []string{"./policy"},
		"set the path to a policy or directory of policies",
	)

	cmd.Flags().BoolVarP(&params.enableTracing, "trace", "t", false, "enable tracing")

	return cmd
}

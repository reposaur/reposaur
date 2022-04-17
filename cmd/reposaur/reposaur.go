package reposaur

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/reposaur/reposaur/pkg/detector"
	"github.com/reposaur/reposaur/pkg/output"
	"github.com/reposaur/reposaur/pkg/sdk"
	"github.com/spf13/cobra"
)

type Params struct {
	namespace    string
	outputFormat string
	policyPaths  []string
}

var cmd = &cobra.Command{
	Use:   "reposaur",
	Short: "Executes a set of Rego policies against the data provided",
	Long:  "Executes a set of Rego policies against the data provided",
}

func NewCommand() *cobra.Command {
	params := Params{}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var input interface{}

		err := json.NewDecoder(os.Stdin).Decode(&input)
		if err != nil {
			return err
		}

		rs, err := sdk.New(cmd.Context(), params.policyPaths)
		if err != nil {
			return err
		}

		namespace := params.namespace

		if namespace == "" {
			namespace, err = detector.DetectNamespace(input)
			if err != nil {
				return err
			}
		}

		report, err := rs.Check(cmd.Context(), namespace, input)
		if err != nil {
			return err
		}

		return writeOutput(report, params.outputFormat, os.Stdout)
	}

	cmd.Flags().StringVarP(
		&params.outputFormat,
		"format", "f", "sarif",
		"report output format (one of 'json' and 'sarif')",
	)

	cmd.Flags().StringVarP(
		&params.namespace,
		"namespace", "n", "",
		"use this namespace",
	)

	cmd.Flags().StringSliceVarP(
		&params.policyPaths,
		"policy", "p", []string{"./policy"},
		"set the path to a policy or directory of policies",
	)

	return cmd
}

func writeOutput(report output.Report, format string, w io.Writer) error {
	switch strings.ToLower(format) {
	case "sarif":
		sarifReport, err := output.NewSarifReport(report)
		if err != nil {
			return err
		}

		return sarifReport.PrettyWrite(w)

	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}

		if _, err := w.Write(data); err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("unknown output format '%s'", format)
}

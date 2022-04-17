package reposaur

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

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

		var data []interface{}

		switch i := input.(type) {
		case map[string]interface{}:
			data = append(data, i)

		case []interface{}:
			for _, d := range i {
				data = append(data, d)
			}
		}

		var (
			wg       = sync.WaitGroup{}
			reportCh = make(chan output.Report, len(data))
		)

		wg.Add(len(data))

		for _, d := range data {
			namespace := params.namespace

			if namespace == "" {
				namespace, err = detector.DetectNamespace(d)
				if err != nil {
					return err
				}
			}

			props, err := detector.DetectReportProperties(namespace, d)
			if err != nil {
				return err
			}

			go func(namespace string, props output.ReportProperties, data interface{}) {
				r, err := rs.Check(cmd.Context(), namespace, data)
				if err != nil {
					panic(err)
				}

				r.Properties = props
				reportCh <- r

				wg.Done()
			}(namespace, props, d)
		}

		wg.Wait()
		close(reportCh)

		var reports []output.Report

		for r := range reportCh {
			reports = append(reports, r)
		}

		return writeOutput(
			reports,
			params.outputFormat,
			os.Stdout,
		)
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

func writeOutput(reports []output.Report, format string, w io.Writer) error {
	format = strings.ToLower(format)

	if format != "json" && format != "sarif" {
		return fmt.Errorf("unknown output format '%s'", format)
	}

	var x []interface{}

	for _, r := range reports {
		if format == "json" {
			x = append(x, r)
			continue
		}

		sarifReport, err := output.NewSarifReport(r)
		if err != nil {
			return err
		}

		x = append(x, sarifReport)
	}

	dec := json.NewEncoder(w)
	dec.SetIndent("", "  ")

	if len(x) == 1 {
		return dec.Encode(x[0])
	}

	return dec.Encode(x)
}

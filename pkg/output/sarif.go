package output

import (
	"strings"

	"github.com/owenrumney/go-sarif/v2/sarif"
)

func NewSarifReport(report Report) (*sarif.Report, error) {
	sr, err := sarif.New(sarif.Version210)
	if err != nil {
		return nil, err
	}

	run := sarif.NewRunWithInformationURI("Reposaur", "https://github.com/reposaur/reposaur")

	run.Properties = sarif.Properties{}
	for k, v := range report.Properties {
		run.Properties[k] = v
	}

	for _, rule := range report.Rules {
		props := sarif.Properties{}

		if len(rule.Tags) > 0 {
			props["tags"] = rule.Tags
		}

		if rule.SecuritySeverity != "" {
			props["security-severity"] = rule.SecuritySeverity
		}

		run.AddRule(rule.UID()).
			WithName(rule.Title).
			WithDescription(rule.Title).
			WithFullDescription(
				sarif.NewMultiformatMessageString(rule.Description).
					WithMarkdown(rule.Description),
			).
			WithMarkdownHelp(rule.Description).
			WithProperties(props)
	}

	for _, result := range report.Results {
		if !result.Passed && !result.Skipped {
			run.AddResult(sarif.NewRuleResult(result.Rule.UID()).
				WithLevel(strings.ToLower(result.Rule.Severity)).
				WithMessage(sarif.NewTextMessage(result.Rule.Title)).
				WithLocations([]*sarif.Location{
					sarif.NewLocation().WithPhysicalLocation(
						sarif.NewPhysicalLocation().
							WithArtifactLocation(
								sarif.NewSimpleArtifactLocation("."),
							),
					),
				}),
			)
		}
	}

	sr.AddRun(run)

	return sr, nil
}

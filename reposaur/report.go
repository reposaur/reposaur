package reposaur

import (
	"fmt"
	"github.com/owenrumney/go-sarif/v2/sarif"
)

const (
	sarifToolName = "Reposaur"
	sarifInfoURI  = "https://github.com/reposaur/reposaur"
)

type Report struct {
	rules []*reportRule
}

func (r Report) SARIF() (*sarif.Report, error) {
	sr, err := sarif.New(sarif.Version210)
	if err != nil {
		return nil, err
	}

	run := sarif.NewRunWithInformationURI(sarifToolName, sarifInfoURI)

	// run.Properties = sarif.Properties{}
	// for k, v := range report.Properties {
	// 	run.Properties[k] = v
	// }

	for _, rule := range r.rules {
		props := sarif.Properties{
			"name":              rule.rule.Name,
			"tags":              rule.rule.Tags,
			"security-severity": fmt.Sprintf("%0.2f", rule.rule.SecuritySeverity),
		}

		run.AddRule(rule.rule.ID).
			WithName(rule.rule.Title).
			WithDescription(rule.rule.Title).
			WithFullDescription(
				sarif.NewMultiformatMessageString(rule.rule.Description).
					WithMarkdown(rule.rule.Description),
			).
			WithMarkdownHelp(rule.rule.Description).
			WithProperties(props)

		if !rule.passed && !rule.skipped {
			run.AddResult(
				sarif.NewRuleResult(rule.rule.ID).
					WithLevel(rule.rule.Kind.Name).
					WithMessage(sarif.NewTextMessage(rule.rule.Title)).
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

type reportRule struct {
	policy  *Policy
	rule    *Rule
	skipped bool
	passed  bool
}

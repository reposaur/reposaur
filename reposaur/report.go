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
			"tags":              rule.rule.Tags,
			"security-severity": fmt.Sprintf("%0.2f", rule.rule.SecuritySeverity),
		}

		uid := fmt.Sprintf("%s/%s/%s", rule.policy.Package, rule.rule.Kind.Name, rule.rule.ID)

		run.AddRule(uid).
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
				sarif.NewRuleResult(uid).
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

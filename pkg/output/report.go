package output

import (
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/ast"
)

type Severity string

const (
	ErrorSeverity   = "error"
	WarningSeverity = "warning"
	NoteSeverity    = "note"
)

var SeverityRuleMap = map[string][]string{
	ErrorSeverity:   {"error", "fail", "violation"},
	WarningSeverity: {"warn"},
	NoteSeverity:    {"note", "info"},
}

var SecuritySeverityMap = map[string]string{
	ErrorSeverity:   "7",
	WarningSeverity: "4",
	NoteSeverity:    "1",
}

type Report struct {
	Rules      map[string]*Rule   `json:"rules"`
	Results    map[string]*Result `json:"results"`
	RuleCount  int                `json:"ruleCount"`
	Properties ReportProperties  `json:"properties"`
}

func (r *Report) AddRule(rule *Rule) {
	r.RuleCount++
	r.Rules[rule.UID()] = rule
}

func (r *Report) AddResult(result *Result) {
	r.Results[result.Rule.UID()] = result
}

type ReportProperties map[string]interface{}

type Result struct {
	Rule    *Rule  `json:"rule"`
	Query   string `json:"query"`
	Skipped bool   `json:"skipped"`
	Passed  bool   `json:"passed"`
}

type Rule struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Kind             string   `json:"kind"`
	Severity         string   `json:"severity"`
	SecuritySeverity string   `json:"security-severity"`
	Description      string   `json:"description"`
	Namespace        string   `json:"namespace"`
	Tags             []string `json:"tags"`
}

func NewRule(namespace string, rule *ast.Rule, as *ast.Annotations) (*Rule, error) {
	headSplit := strings.SplitN(rule.Head.Name.String(), "_", 2)

	if len(headSplit) != 2 {
		return nil, fmt.Errorf("new rule: parse id: invalid rule name: %s", rule.Head.Name.String())
	}

	var (
		kind     = headSplit[0]
		id       = headSplit[1]
		severity = ""
	)

	for sev, kinds := range SeverityRuleMap {
		for _, k := range kinds {
			if k == kind {
				severity = sev
				break
			}
		}

		if severity != "" {
			break
		}
	}

	if severity == "" {
		return nil, fmt.Errorf("new rule: could not find severity for %s", kind)
	}

	r := Rule{
		ID:               id,
		Title:            id,
		Kind:             kind,
		Severity:         severity,
		SecuritySeverity: SecuritySeverityMap[severity],
		Namespace:        namespace,
	}

	if as != nil {
		if as.Title != "" {
			r.Title = as.Title
		} else {
			r.Title = r.ID
		}

		if as.Description != "" {
			r.Description = as.Description
		} else {
			r.Description = r.Title
		}

		if tags, ok := as.Custom["tags"]; ok {
			for _, t := range tags.([]interface{}) {
				r.Tags = append(r.Tags, t.(string))
			}
		}

		if secSev, ok := as.Custom["security-severity"]; ok {
			r.SecuritySeverity = fmt.Sprintf("%v", secSev)
		}
	}

	return &r, nil
}

func (r Rule) CausesFailure() bool {
	return r.Severity == ErrorSeverity
}

func (r Rule) UID() string {
	return fmt.Sprintf("%s/%s/%s", r.Namespace, r.Kind, r.ID)
}

func MergeReports(reports []Report) Report {
	report := Report{
		Rules:   make(map[string]*Rule),
		Results: make(map[string]*Result),
	}

	for _, r := range reports {
		report.RuleCount += r.RuleCount

		for k, v := range r.Rules {
			report.Rules[k] = v
		}

		for k, v := range r.Results {
			report.Results[k] = v
		}
	}

	return report
}

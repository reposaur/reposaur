package reposaur

import (
	"errors"
	"github.com/open-policy-agent/opa/ast"
	"strings"
)

var ruleKinds = []*RuleKind{
	{
		Name:             "error",
		Aliases:          []string{"violation", "fail"},
		SecuritySeverity: 7.0,
	},
	{
		Name:             "warning",
		Aliases:          []string{"warn"},
		SecuritySeverity: 4.0,
	},
	{
		Name:             "note",
		Aliases:          []string{"info"},
		SecuritySeverity: 1.0,
	},
}

var errSkipRule = errors.New("skip rule")

type Policy struct {
	Metadata

	Package string  `json:"package"`
	Rules   []*Rule `json:"rules"`
}

type Rule struct {
	Metadata

	ID   string
	Kind *RuleKind

	name   string
	schema ast.Ref
}

func newRule(r *ast.Rule) (rule *Rule, err error) {
	rule = &Rule{}

	rule.ID, rule.Kind, err = parseRuleName(r)
	if err != nil {
		return nil, err
	}

	for _, a := range r.Module.Annotations {
		if a.Scope == "rule" && a.GetTargetPath().String() == r.Ref().String() {
			rule.Metadata = newMetadata(a)

			for _, schema := range a.Schemas {
				if schema.Path.String() != "input" {
					continue
				}
				rule.schema = schema.Schema.Copy()
				break
			}
			break
		}
	}

	if rule.schema == nil {
		return nil, errSkipRule
	}

	if rule.SecuritySeverity == 0 {
		rule.SecuritySeverity = rule.Kind.SecuritySeverity
	}

	return rule, nil
}

func parseRuleName(rule *ast.Rule) (string, *RuleKind, error) {
	headSplit := strings.SplitN(rule.Head.Name.String(), "_", 2)

	if len(headSplit) != 2 {
		return "", nil, errSkipRule
	}

	kind := ruleKindForString(headSplit[0])
	if kind == nil {
		return "", nil, errSkipRule
	}

	return rule.Head.Name.String(), kind, nil
}

type RuleKind struct {
	// Name correlates with SARIF level property
	Name string
	// Aliases are a list of prefixes a rule can have to match this kind
	Aliases          []string
	SecuritySeverity float64
}

func ruleKindForString(str string) *RuleKind {
	for _, kind := range ruleKinds {
		if str == kind.Name || contains(kind.Aliases, str) {
			return kind
		}
	}
	return nil
}

type Author struct {
	Name  string
	Email string
}

type Metadata struct {
	Title            string   `json:"title,omitempty"`
	Description      string   `json:"description,omitempty"`
	Authors          []Author `json:"authors,omitempty"`
	Organizations    []string `json:"organizations,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	SecuritySeverity float64  `json:"securitySeverity,omitempty"`
}

func newMetadata(a *ast.Annotations) Metadata {
	md := Metadata{}

	if a == nil {
		return md
	}

	md.Title = a.Title
	md.Description = a.Description
	md.Organizations = a.Organizations
	md.Tags = parseTagsAnnotation(a)

	if secSev := parseSecuritySeverityAnnotation(a); secSev != 0 {
		md.SecuritySeverity = secSev
	}

	for _, author := range a.Authors {
		md.Authors = append(md.Authors, Author{
			Name:  author.Name,
			Email: author.Email,
		})
	}

	return md
}

func parseTagsAnnotation(a *ast.Annotations) []string {
	tags, ok := a.Custom["tags"]
	if !ok {
		return nil
	}

	switch v := tags.(type) {
	case []any:
		var parsedTags []string
		for _, t := range v {
			if vv, ok := t.(string); ok {
				parsedTags = append(parsedTags, vv)
			}
		}
		return parsedTags
	case string:
		return []string{v}
	}

	return nil
}

func parseSecuritySeverityAnnotation(a *ast.Annotations) float64 {
	secSev, ok := a.Custom["security-severity"]
	if !ok {
		return 0
	}

	switch v := secSev.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	}

	return 0
}

func contains(slice []string, target string) bool {
	for _, str := range slice {
		if str == target {
			return true
		}
	}
	return false
}

func parsePackageName(mod *ast.Module) string {
	return strings.TrimPrefix(mod.Package.Path.String(), "data.")
}

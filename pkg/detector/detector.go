package detector

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"text/template"

	"github.com/reposaur/reposaur/pkg/output"
)

// ErrUnknownNamespace happens when DetectNamespace
// fails to detect a valid namespace.
var ErrUnknownNamespace = errors.New("failed to detect namespace from data")

var ErrUnknownReportProperties = errors.New("failed to detect report properties from namespace and data")

var namespaceToKeysMap = map[string][]string{
	"issue":        {"reactions", "closed_by"},
	"organization": {"login", "members_url"},
	"pull_request": {"base", "head"},
	"repository":   {"owner", "full_name"},
	"user":         {"login", "hireable"},
}

var namespaceToReportPropertiesMap = map[string]string{
	"issue": `
		{
			"id": {{.id}},
			"number": {{.number}}
		}
	`,
	"organization": `
		{
			"login": "{{.login}}",
			"name": "{{.name}}"
		}
	`,
	"pull_request": `
		{
			"id": {{.id}},
			"number": {{.number}}
		}
	`,
	"repository": `
		{
			"owner": "{{.owner.login}}",
			"repo": "{{.name}}",
			"default_branch": "{{.default_branch}}"
		}
	`,
	"user": `
		{
			"login": "{{.login}}",
			"name": "{{.name}}"
		}
	`,
}

// DetectNamespace will attempt to detect some data's
// namespace based on it's keys.
//
// Returns an error of type ErrUnknownNamespace if data
// is not a map or if it can't detect a valid namespace.
func DetectNamespace(data interface{}) (string, error) {
	val := reflect.ValueOf(data)

	if val.Kind() != reflect.Map {
		return "", ErrUnknownNamespace
	}

	for n, keys := range namespaceToKeysMap {
		var matches int

		for _, k := range keys {
			for _, dk := range val.MapKeys() {
				if k == dk.String() {
					matches++
				}
			}

			if matches == len(keys) {
				return n, nil
			}
		}
	}

	return "", ErrUnknownNamespace
}

func DetectReportProperties(namespace string, data interface{}) (output.ReportProperties, error) {
	if tplStr, ok := namespaceToReportPropertiesMap[namespace]; ok {
		tpl := template.Must(template.New(namespace).Parse(tplStr))

		buf := &bytes.Buffer{}
		if err := tpl.Execute(buf, data); err != nil {
			return output.ReportProperties{}, err
		}

		props := output.ReportProperties{}
		dec := json.NewDecoder(buf)

		if err := dec.Decode(&props); err != nil {
			return output.ReportProperties{}, err
		}

		return props, nil
	}

	return output.ReportProperties{}, ErrUnknownReportProperties
}

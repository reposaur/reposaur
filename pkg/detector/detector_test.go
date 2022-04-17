package detector_test

import (
	"reflect"
	"testing"

	"github.com/reposaur/reposaur/pkg/detector"
	"github.com/reposaur/reposaur/pkg/output"
)

func TestDetectIssueNamespace(t *testing.T) {
	data := map[string]interface{}{
		"reactions": "some",
		"closed_by": "some",
	}

	ns, err := detector.DetectNamespace(data)
	if err != nil {
		t.Fatal(err)
	}

	expected := "issue"

	if ns != expected {
		t.Errorf("expected namespace to be %s, got '%s'", expected, ns)
	}
}

func TestDetectOrganizationNamespace(t *testing.T) {
	data := map[string]interface{}{
		"login":       "reposaur",
		"members_url": "https://github.com",
	}

	ns, err := detector.DetectNamespace(data)
	if err != nil {
		t.Fatal(err)
	}

	expected := "organization"

	if ns != expected {
		t.Errorf("expected namespace to be %s, got '%s'", expected, ns)
	}
}

func TestDetectPullRequestNamespace(t *testing.T) {
	data := map[string]interface{}{
		"base": "some",
		"head": "some",
	}

	ns, err := detector.DetectNamespace(data)
	if err != nil {
		t.Fatal(err)
	}

	expected := "pull_request"

	if ns != expected {
		t.Errorf("expected namespace to be %s, got '%s'", expected, ns)
	}
}

func TestDetectRepositoryNamespace(t *testing.T) {
	data := map[string]interface{}{
		"owner":     "reposaur",
		"full_name": "reposaur/test",
	}

	ns, err := detector.DetectNamespace(data)
	if err != nil {
		t.Fatal(err)
	}

	expected := "repository"

	if ns != expected {
		t.Errorf("expected namespace to be %s, got '%s'", expected, ns)
	}
}

func TestDetectUserNamespace(t *testing.T) {
	data := map[string]interface{}{
		"login":    "crqra",
		"hireable": true,
	}

	ns, err := detector.DetectNamespace(data)
	if err != nil {
		t.Fatal(err)
	}

	expected := "user"

	if ns != expected {
		t.Errorf("expected namespace to be %s, got '%s'", expected, ns)
	}
}

func TestDetectIssueReportProperties(t *testing.T) {
	data := map[string]interface{}{
		"id":     123,
		"number": 123,
	}

	props, err := detector.DetectReportProperties("issue", data)
	if err != nil {
		t.Fatal(err)
	}

	expected := output.ReportProperties{
		"id":     float64(123),
		"number": float64(123),
	}

	if !reflect.DeepEqual(expected, props) {
		t.Errorf("expected report properties to be %v, got %v", expected, props)
	}
}

func TestDetectOrganizationReportProperties(t *testing.T) {
	data := map[string]interface{}{
		"login": "reposaur",
		"name":  "Reposaur",
	}

	props, err := detector.DetectReportProperties("organization", data)
	if err != nil {
		t.Fatal(err)
	}

	expected := output.ReportProperties{
		"login": "reposaur",
		"name":  "Reposaur",
	}

	if !reflect.DeepEqual(expected, props) {
		t.Errorf("expected report properties to be %v, got %v", expected, props)
	}
}

func TestDetectPullRequestReportProperties(t *testing.T) {
	data := map[string]interface{}{
		"id":     123,
		"number": 123,
	}

	props, err := detector.DetectReportProperties("pull_request", data)
	if err != nil {
		t.Fatal(err)
	}

	expected := output.ReportProperties{
		"id":     float64(123),
		"number": float64(123),
	}

	if !reflect.DeepEqual(expected, props) {
		t.Errorf("expected report properties to be %v, got %v", expected, props)
	}
}

func TestDetectRepositoryReportProperties(t *testing.T) {
	data := map[string]interface{}{
		"owner": map[string]interface{}{
			"login": "reposaur",
		},
		"name":           "reposaur",
		"default_branch": "main",
	}

	props, err := detector.DetectReportProperties("repository", data)
	if err != nil {
		t.Fatal(err)
	}

	expected := output.ReportProperties{
		"owner":          "reposaur",
		"repo":           "reposaur",
		"default_branch": "main",
	}

	if !reflect.DeepEqual(expected, props) {
		t.Errorf("expected report properties to be %v, got %v", expected, props)
	}
}

func TestDetectUserReportProperties(t *testing.T) {
	data := map[string]interface{}{
		"login": "reposaur",
		"name":  "Reposaur",
	}

	props, err := detector.DetectReportProperties("user", data)
	if err != nil {
		t.Fatal(err)
	}

	expected := output.ReportProperties{
		"login": "reposaur",
		"name":  "Reposaur",
	}

	if !reflect.DeepEqual(expected, props) {
		t.Errorf("expected report properties to be %v, got %v", expected, props)
	}
}

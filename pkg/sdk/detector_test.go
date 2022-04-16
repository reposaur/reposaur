package sdk_test

import (
	"testing"

	"github.com/reposaur/reposaur/pkg/sdk"
)

func TestDetectIssueNamespace(t *testing.T) {
	data := map[string]interface{}{
		"reactions": "some",
		"closed_by": "some",
	}

	ns, err := sdk.DetectNamespace(data)
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

	ns, err := sdk.DetectNamespace(data)
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

	ns, err := sdk.DetectNamespace(data)
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

	ns, err := sdk.DetectNamespace(data)
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

	ns, err := sdk.DetectNamespace(data)
	if err != nil {
		t.Fatal(err)
	}

	expected := "user"

	if ns != expected {
		t.Errorf("expected namespace to be %s, got '%s'", expected, ns)
	}
}

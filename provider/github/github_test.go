package github_test

import (
	"testing"

	"github.com/reposaur/reposaur/provider"
	"github.com/reposaur/reposaur/provider/github"
)

func TestDeriveNamespace(t *testing.T) {
	gh := github.New(nil)

	testData := map[provider.Namespace]map[string]any{
		github.IssueNamespace: {
			"reactions": ":+1:",
			"closed_by": "crqra",
		},
		github.OrganizationNamespace: {
			"login":       "reposaur",
			"members_url": "https://reposaur.com",
		},
		github.PullRequestNamespace: {
			"base": "main",
			"head": "feat",
		},
		github.RepositoryNamespace: {
			"owner":     "reposaur",
			"full_name": "reposaur/reposaur",
		},
		github.UserNamespace: {
			"login":    "crqra",
			"hireable": true,
		},
	}

	for expected, data := range testData {
		namespace, err := gh.DeriveNamespace(data)
		if err != nil {
			t.Fatalf("testing %s: %s", expected, err)
		}

		if namespace != expected {
			t.Fatalf("expected namespace to be '%s' got '%s'", expected, namespace)
		}
	}
}

package gitlab_test

import (
	"testing"

	"github.com/reposaur/reposaur/provider"
	"github.com/reposaur/reposaur/provider/gitlab"
)

func TestDeriveNamespace(t *testing.T) {
	gl := gitlab.NewProvider()

	testData := map[provider.Namespace]map[string]any{
		gitlab.GroupNamespace: {
			"projects":                []any{},
			"subgroup_creation_level": "maintainer",
		},
		gitlab.MergeRequestNamespace: {
			"merge_user":   "",
			"merge_status": "",
			"reference":    "",
		},
		gitlab.ProjectNamespace: {
			"namespace": map[string]any{
				"name": "test",
			},
			"name_with_namespace": "test / test",
		},
		gitlab.UserNamespace: {
			"bio":      "",
			"bot":      "",
			"pronouns": "",
		},
	}

	for expected, data := range testData {
		namespace, err := gl.DeriveNamespace(data)
		if err != nil {
			t.Fatalf("testing %s: %s", expected, err)
		}

		if namespace != expected {
			t.Fatalf("expected namespace to be '%s' got '%s'", expected, namespace)
		}
	}
}

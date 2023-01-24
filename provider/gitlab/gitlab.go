package gitlab

import (
	"github.com/reposaur/reposaur/provider"
)

const (
	GroupNamespace        provider.Namespace = "gitlab.group"
	MergeRequestNamespace provider.Namespace = "gitlab.merge_request"
	ProjectNamespace      provider.Namespace = "gitlab.project"
	UserNamespace         provider.Namespace = "gitlab.user"
)

type GitLab struct {
	dataDeriver *DataDeriver
	builtins    []provider.Builtin
}

func NewProvider() *GitLab {
	return &GitLab{
		dataDeriver: &DataDeriver{
			namespaceToKeys: map[provider.Namespace][]string{
				GroupNamespace:        {"projects", "subgroup_creation_level"},
				MergeRequestNamespace: {"merge_user", "merge_status", "reference"},
				ProjectNamespace:      {"namespace", "name_with_namespace"},
				UserNamespace:         {"bio", "bot", "pronouns"},
			},
		},
	}
}

func (gl GitLab) DeriveNamespace(data map[string]any) (provider.Namespace, error) {
	return gl.dataDeriver.DeriveNamespace(data)
}

func (gl GitLab) DeriveProperties(namespace provider.Namespace, data map[string]any) (map[string]any, error) {
	return gl.dataDeriver.DeriveProperties(namespace, data)
}

func (gl GitLab) Builtins() []provider.Builtin {
	return gl.builtins
}

type DataDeriver struct {
	namespaceToKeys map[provider.Namespace][]string
}

func (d DataDeriver) DeriveNamespace(data map[string]any) (provider.Namespace, error) {
	for namespace, keys := range d.namespaceToKeys {
		var matches int

		for _, key := range keys {
			for dataKey := range data {
				if key == dataKey {
					matches++
				}
			}

			if matches == len(keys) {
				return namespace, nil
			}
		}
	}

	return "", provider.ErrNonDerivable
}

func (d DataDeriver) DeriveProperties(namespace provider.Namespace, data map[string]any) (map[string]any, error) {
	return nil, provider.ErrNonDerivable
}

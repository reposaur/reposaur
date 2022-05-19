package github

import (
	"github.com/reposaur/reposaur/provider"
	"github.com/reposaur/reposaur/provider/github/internal/builtin"
	"github.com/reposaur/reposaur/provider/github/internal/client"
)

const (
	IssueNamespace        provider.Namespace = "github.issue"
	OrganizationNamespace provider.Namespace = "github.organization"
	PullRequestNamespace  provider.Namespace = "github.pull_request"
	RepositoryNamespace   provider.Namespace = "github.repository"
	UserNamespace         provider.Namespace = "github.user"
)

type GitHub struct {
	client      client.Client
	dataDeriver *DataDeriver
	builtins    []provider.Builtin
}

func New(c client.Client) *GitHub {
	if c == nil {
		c = client.DefaultClient
	}

	c.Client().Transport = client.NewThrottleTransport(
		c.Client().Transport,
	)

	return &GitHub{
		client: c,
		builtins: []provider.Builtin{
			&builtin.GraphQL{Client: c},
			&builtin.Request{Client: c},
		},
		dataDeriver: &DataDeriver{
			namespaceToKeys: map[provider.Namespace][]string{
				IssueNamespace:        {"reactions", "closed_by"},
				OrganizationNamespace: {"login", "members_url"},
				PullRequestNamespace:  {"base", "head"},
				RepositoryNamespace:   {"owner", "full_name"},
				UserNamespace:         {"login", "hireable"},
			},
		},
	}
}

func (gh GitHub) DeriveNamespace(data map[string]any) (provider.Namespace, error) {
	return gh.dataDeriver.DeriveNamespace(data)
}

func (gh GitHub) DeriveProperties(namespace provider.Namespace, data map[string]any) (map[string]any, error) {
	return gh.dataDeriver.DeriveProperties(namespace, data)
}

func (gh GitHub) Builtins() []provider.Builtin {
	return gh.builtins
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
	switch namespace {
	case IssueNamespace, PullRequestNamespace:
		props := map[string]any{}

		if id, ok := data["id"]; ok {
			props["id"] = id
		}

		if nr, ok := data["number"]; ok {
			props["number"] = nr
		}

		return props, nil

	case OrganizationNamespace, UserNamespace:
		props := map[string]any{}

		if login, ok := data["login"]; ok {
			props["login"] = login
		}

		if name, ok := data["name"]; ok {
			props["name"] = name
		}

		return props, nil

	case RepositoryNamespace:
		props := map[string]any{}

		if owner, ok := data["owner"].(map[string]any); ok {
			if login, ok := owner["login"]; ok {
				props["owner"] = login
			}
		}

		if name, ok := data["name"]; ok {
			props["repo"] = name
		}

		if defaultBranch, ok := data["default_branch"]; ok {
			props["default_branch"] = defaultBranch
		}

		return props, nil
	}

	return nil, provider.ErrNonDerivable
}

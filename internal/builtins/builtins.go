package builtins

import (
	"github.com/open-policy-agent/opa/rego"
	"github.com/reposaur/reposaur/pkg/github"
)

func RegisterBuiltins(client *github.Client) {
	rego.RegisterBuiltin2(&GitHubRequestBuiltin, GitHubRequestBuiltinImpl(client))
	rego.RegisterBuiltin2(&GitHubGraphQLBuiltin, GitHubGraphQLBuiltinImpl(client))
}

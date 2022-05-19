package builtins

import (
	"net/http"

	"github.com/open-policy-agent/opa/rego"
)

func RegisterBuiltins(client *http.Client) {
	rego.RegisterBuiltin2(&GitHubGraphQLBuiltin, GitHubGraphQLBuiltinImpl(client))
}

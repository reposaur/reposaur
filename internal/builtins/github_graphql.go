package builtins

import (
	"encoding/json"
	"net/http"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/reposaur/reposaur/pkg/github"
)

var GitHubGraphQLBuiltin = rego.Function{
	Name: "github.graphql",
	Decl: types.NewFunction(
		types.Args(
			types.S,
			types.NewObject(nil, types.NewDynamicProperty(types.S, types.A)),
		),
		types.A,
	),
	Memoize: true,
}

func GitHubGraphQLBuiltinImpl(client *github.Client) func(bctx rego.BuiltinContext, op1, op2 *ast.Term) (*ast.Term, error) {
	return func(bctx rego.BuiltinContext, op1, op2 *ast.Term) (*ast.Term, error) {
		var query string
		var variables map[string]interface{}

		if err := ast.As(op1.Value, &query); err != nil {
			return nil, err
		} else if err := ast.As(op2.Value, &variables); err != nil {
			return nil, err
		}

		body := map[string]interface{}{
			"query":     query,
			"variables": variables,
		}

		req, err := client.NewRequest(http.MethodPost, "/graphql", body)
		if err != nil {
			return nil, err
		}

		finalResp := GitHubResponse{}
		resp, err := client.Do(req)
		if err != nil {
			finalResp.Error = err.Error()
		}
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&finalResp.Body); err != nil {
			return nil, err
		}

		finalResp.StatusCode = resp.StatusCode

		val, err := ast.InterfaceToValue(finalResp)
		if err != nil {
			return nil, err
		}

		return ast.NewTerm(val), nil
	}
}

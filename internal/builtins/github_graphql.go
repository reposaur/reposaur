package builtins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
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

func GitHubGraphQLBuiltinImpl(client *http.Client) func(bctx rego.BuiltinContext, op1, op2 *ast.Term) (*ast.Term, error) {
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

		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(body); err != nil {
			return nil, err
		}

		req, err := http.NewRequest(http.MethodPost, "/graphql", buf)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", "reposaur")
		req.Header.Set("Content-Type", "application/json")

		finalResp := GitHubResponse{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&finalResp.Body); err != nil {
			return nil, err
		}

		finalResp.StatusCode = resp.StatusCode

		if finalResp.StatusCode == http.StatusForbidden {
			b := finalResp.Body.(map[string]interface{})
			return nil, fmt.Errorf("forbidden: %s", b["message"])
		}

		val, err := ast.InterfaceToValue(finalResp)
		if err != nil {
			return nil, err
		}

		return ast.NewTerm(val), nil
	}
}

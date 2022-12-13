package builtin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/reposaur/reposaur/provider/github/client"
)

type GraphQL struct {
	Client *client.Client
}

func (gql GraphQL) Func() *rego.Function {
	return &rego.Function{
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
}

func (gql GraphQL) Impl(_ rego.BuiltinContext, terms []*ast.Term) (*ast.Term, error) {
	req, err := gql.argsToRequest(terms)
	if err != nil {
		return nil, err
	}

	resp, err := gql.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var finalResp response

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

func (gql GraphQL) argsToRequest(terms []*ast.Term) (*retryablehttp.Request, error) {
	// FIXME: Function receives 2 arguments but terms includes one additional at last index
	if len(terms) != 3 {
		return nil, fmt.Errorf("wrong number of arguments, expected 2 got %d", len(terms)-1)
	}

	var (
		query string
		vars  map[string]any
	)

	if err := ast.As(terms[0].Value, &query); err != nil {
		return nil, err
	}

	if err := ast.As(terms[1].Value, &vars); err != nil {
		return nil, err
	}

	body := map[string]any{
		"query":     query,
		"variables": vars,
	}

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(body); err != nil {
		return nil, err
	}

	return gql.Client.NewRequest(http.MethodPost, "/graphql", buf)
}

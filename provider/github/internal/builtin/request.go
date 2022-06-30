package builtin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/reposaur/reposaur/provider/github/client"
)

const builtinInputArgs = 2

type Request struct {
	Client *client.Client
}

// The function performs an http request to the Github API and is intend to work similarly to Github API v3.
// Please check https://docs.github.com/en/rest for more information.
// First term will relate to an unparsed request string that looks like the following:
// "GET /repos/{owner}/{repo}/branches/{branch}/protection"
// The second term relates to the request path values being sent to the api. It is defined as an object and will
// map the strings in between brackets to their provided value in this object
// {
//	"owner": "reposaur",
//	"repo":  "reposaur",
//	"branch": "main",
// }.
func (r Request) Func() *rego.Function {
	return &rego.Function{
		Name: "github.request",
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

func (r Request) Impl(ctx rego.BuiltinContext, terms []*ast.Term) (*ast.Term, error) {
	req, err := r.argsToRequest(terms)
	if err != nil {
		return nil, err
	}

	resp, err := r.Client.Do(req)
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
		b := finalResp.Body.(map[string]any)
		return nil, fmt.Errorf("forbidden: %s", b["message"])
	}

	val, err := ast.InterfaceToValue(finalResp)
	if err != nil {
		return nil, err
	}

	return ast.NewTerm(val), nil
}

func (r Request) argsToRequest(terms []*ast.Term) (*retryablehttp.Request, error) {
	// FIXME: Function receives 2 arguments but terms includes one additional at last index
	if len(terms) != 3 {
		return nil, fmt.Errorf("wrong number of arguments, expected 2 got %d", len(terms)-1)
	}

	var fullPath string
	if err := ast.As(terms[0].Value, &fullPath); err != nil {
		return nil, err
	}

	var data map[string]any
	if err := ast.As(terms[1].Value, &data); err != nil {
		return nil, err
	}

	method, path, err := splitPath(fullPath)
	if err != nil {
		return nil, err
	}

	if method != http.MethodGet {
		return nil, fmt.Errorf("only GET requests are supported, got '%s'", method)
	}

	u, err := buildRequestUrl(path, data)
	if err != nil {
		return nil, err
	}

	return r.Client.NewRequest(method, u, nil)
}

func buildRequestUrl(path string, data map[string]any) (string, error) {
	pathParams := parsePathParams(path)

	if len(pathParams) != 0 {
		// replace path parameters with values
		for _, p := range pathParams {
			v, err := valueToString(data[p])
			if err != nil {
				return "", err
			}

			path = strings.Replace(path, "{"+p+"}", v, 1)
			delete(data, p)
		}
	}

	qs := url.Values{}
	// adding remaining data values as query strings
	for k, v := range data {
		v, err := valueToString(v)
		if err != nil {
			return "", err
		}

		qs.Add(k, v)
		delete(data, k)
	}

	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	u.RawQuery = qs.Encode()

	return u.String(), err
}

func splitPath(p string) (string, string, error) {
	pathParts := strings.Split(p, " ")

	if len(pathParts) != builtinInputArgs {
		return "", "", fmt.Errorf("wrong number of parts in path, expected 2 got %d", len(pathParts))
	}

	var (
		method = strings.ToUpper(pathParts[0])
		path   = pathParts[1]
	)

	return method, path, nil
}

func parsePathParams(path string) []string {
	regex := regexp.MustCompile(`{[a-z]+}`)
	matches := regex.FindAllString(path, -1)

	var params []string
	for _, v := range matches {
		p := strings.Replace(v, "{", "", 1)
		p = strings.Replace(p, "}", "", 1)
		params = append(params, p)
	}

	return params
}

func valueToString(v any) (string, error) {
	switch tv := v.(type) {
	case string:
		return tv, nil
	case json.Number:
		return tv.String(), nil
	case int:
		return strconv.Itoa(tv), nil
	case int64:
		return strconv.Itoa(int(tv)), nil
	}

	return "", fmt.Errorf("parse error: can't parse '%v' to string", v)
}

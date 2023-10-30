package reposaur

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/metrics"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/reposaur/reposaur/provider"
	"io/fs"
	"os"
	"reflect"
	"strings"
)

type Reposaur struct {
	policies  map[string]*Policy
	providers map[string]provider.Provider

	// rego specific
	fs           billy.Filesystem
	loader       loader.FileLoader
	metrics      metrics.Metrics
	schemas      *ast.SchemaSet
	compiler     *ast.Compiler
	capabilities *ast.Capabilities
}

// New creates a new Reposaur instance.
func New(opts ...Option) (*Reposaur, error) {
	r := &Reposaur{
		providers:    map[string]provider.Provider{},
		policies:     map[string]*Policy{},
		metrics:      metrics.New(),
		schemas:      ast.NewSchemaSet(),
		capabilities: ast.CapabilitiesForThisVersion(),
	}

	for _, opt := range opts {
		opt(r)
	}

	if r.providers == nil {
		return nil, errors.New("at least one provider must be specified")
	}

	if r.fs == nil {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		r.fs = osfs.New(wd)
	}

	for _, p := range r.providers {
		pkey := ast.Ref{ast.SchemaRootDocument, ast.StringTerm(providerName(p))}
		for _, s := range p.Schemas() {
			key := pkey.Copy().Append(ast.StringTerm(s.Name()))
			r.schemas.Put(key, s.Raw())
		}

		// TODO: implement builtins
		// r.capabilities.Builtins = append(r.capabilities.Builtins, p.Builtins()...)
	}

	r.loader = loader.NewFileLoader().
		WithFS(billyFS{r.fs}).
		WithMetrics(r.metrics).
		WithProcessAnnotation(true).
		WithCapabilities(r.capabilities)

	r.compiler = ast.NewCompiler().
		WithStrict(true).
		WithEnablePrintStatements(true).
		WithMetrics(r.metrics).
		WithCapabilities(r.capabilities).
		WithSchemas(r.schemas)

	return r, nil
}

// Policies returns all that policies that have been loaded.
func (r *Reposaur) Policies() map[string]*Policy {
	return r.policies
}

// LoadPolicies loads all the policies available in the paths. A path
// can be the path to a single .rego file or a directory.
func (r *Reposaur) LoadPolicies(paths ...string) error {
	res, err := r.loader.Filtered(paths, func(_ string, info fs.FileInfo, _ int) bool {
		return !info.IsDir() && !strings.HasSuffix(info.Name(), bundle.RegoExt)
	})
	if err != nil {
		return err
	}

	for _, mod := range res.ParsedModules() {
		pkgName := parsePackageName(mod)

		policy, ok := r.policies[pkgName]
		if !ok {
			policy = &Policy{ID: genID(pkgName), Package: pkgName}
			for _, a := range mod.Annotations {
				if a.Scope == "package" {
					policy.Metadata = newMetadata(a)
				}
			}
		}

		for _, r := range mod.Rules {
			rule, err := newRule(r)
			if err != nil {
				if errors.Is(err, errSkipRule) {
					continue
				}
				return err
			}
			policy.Rules = append(policy.Rules, rule)
		}

		r.policies[pkgName] = policy
	}

	r.compiler.Compile(res.ParsedModules())
	if r.compiler.Failed() {
		return r.compiler.Errors
	}

	return nil
}

// Eval evaluates the loaded policies against the given input. A schema
// will be derived from the input and rules that specify that schema are evaluated.
func (r *Reposaur) Eval(ctx context.Context, input any) (*Report, error) {
	inputSchema := r.inputSchema(input)
	if inputSchema == nil {
		return nil, errors.New("unknown schema for input")
	}

	report := &Report{}

	for _, policy := range r.policies {
		for _, rule := range policy.Rules {
			// ignore rules that don't specify the `schema` annotation
			if rule.schema == nil || !rule.schema.Equal(inputSchema) {
				continue
			}

			reportRule := &reportRule{policy: policy, rule: rule}

			// check if rule should be skipped
			skipQuery := fmt.Sprintf("data.%s.skip[_][_] == %q", policy.Package, rule.Name)
			rs, err := r.doEval(ctx, skipQuery, input)
			if err != nil {
				return nil, err
			}

			if len(rs) > 0 {
				reportRule.skipped = true
				report.rules = append(report.rules, reportRule)
				continue
			}

			query := fmt.Sprintf("data.%s.%s", policy.Package, rule.Name)
			rs, err = r.doEval(ctx, query, input)
			if err != nil {
				return nil, err
			}

			reportRule.passed = len(rs) == 0
			report.rules = append(report.rules, reportRule)
		}
	}

	return report, nil
}

func (r *Reposaur) doEval(ctx context.Context, query string, input any) (rego.ResultSet, error) {
	ri := rego.New(
		rego.Query(query),
		rego.Input(input),
		rego.Compiler(r.compiler),
		rego.Metrics(r.metrics),
		rego.Schemas(r.schemas),
		rego.Capabilities(r.capabilities),
		rego.Trace(true),
		rego.EnablePrintStatements(true),
		rego.PrintHook(topdown.NewPrintHook(os.Stderr)),
	)
	return ri.Eval(ctx)
}

func (r *Reposaur) inputSchema(input any) ast.Ref {
	for _, p := range r.providers {
		for _, s := range p.Schemas() {
			if err := s.Validate(input); err == nil {
				return ast.Ref{
					ast.SchemaRootDocument,
					ast.StringTerm(providerName(p)),
					ast.StringTerm(s.Name()),
				}
			}
		}
	}
	return nil
}

func providerName(p provider.Provider) string {
	return strings.ToLower(reflect.TypeOf(p).Elem().Name())
}

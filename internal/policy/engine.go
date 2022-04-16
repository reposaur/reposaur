package policy

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/reposaur/reposaur/pkg/output"
)

type Engine struct {
	modules  map[string]*ast.Module
	compiler *ast.Compiler
}

func Load(ctx context.Context, policyPaths []string) (*Engine, error) {
	policies, err := allRegos(policyPaths)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	} else if len(policies.Modules) == 0 {
		return nil, fmt.Errorf("no policies found in %v", policyPaths)
	}

	modules := policies.ParsedModules()
	compiler := ast.NewCompiler().WithEnablePrintStatements(true)

	compiler.Compile(modules)

	if compiler.Failed() {
		return nil, fmt.Errorf("compiler: %w", compiler.Errors)
	}

	engine := Engine{
		modules:  modules,
		compiler: compiler,
	}

	return &engine, nil
}

// Namespaces returns all of the namespaces in the engine.
func (e *Engine) Namespaces() []string {
	var namespaces []string
	for _, module := range e.Modules() {
		namespace := strings.Replace(module.Package.Path.String(), "data.", "", 1)
		for _, ns := range namespaces {
			if ns == namespace {
				continue
			}
		}

		namespaces = append(namespaces, namespace)
	}

	return namespaces
}

// Compiler returns the compiler from the loaded policies.
func (e *Engine) Compiler() *ast.Compiler {
	return e.compiler
}

// Modules returns the modules from the loaded policies.
func (e *Engine) Modules() map[string]*ast.Module {
	return e.modules
}

func (e *Engine) Check(ctx context.Context, namespace string, input interface{}) (output.Report, error) {
	report, err := e.check(ctx, namespace, input)
	if err != nil {
		return output.Report{}, fmt.Errorf("check: %w", err)
	}

	return report, nil
}

func (e *Engine) check(ctx context.Context, namespace string, input interface{}) (output.Report, error) {
	report := output.Report{
		Rules:   map[string]*output.Rule{},
		Results: map[string]*output.Result{},
	}

	for _, mod := range e.Modules() {
		currNamespace := strings.TrimLeft(mod.Package.Path.String(), "data.")
		if currNamespace != namespace {
			continue
		}

		for _, r := range mod.Rules {
			var annotations *ast.Annotations
			for _, a := range mod.Annotations {
				if a.Scope == "rule" && a.GetTargetPath().String() == r.Path().String() {
					annotations = a
				}
			}

			rule, err := output.NewRule(namespace, r, annotations)
			if err != nil {
				continue
			}

			report.AddRule(rule)
		}
	}

	for _, rule := range report.Rules {
		var result *output.Result

		result, err := e.querySkip(ctx, rule, input)
		if err != nil {
			return output.Report{}, fmt.Errorf("query skip rule: %s: %w", rule.UID(), err)
		}

		if !result.Skipped {
			result, err = e.queryRule(ctx, rule, input)
			if err != nil {
				return output.Report{}, fmt.Errorf("query rule: %s: %w", rule.UID(), err)
			}
		}

		report.AddResult(result)
	}

	return report, nil
}

func (e Engine) queryRule(ctx context.Context, rule *output.Rule, input interface{}) (*output.Result, error) {
	query := fmt.Sprintf("data.%s.%s_%s", rule.Namespace, rule.Kind, rule.ID)
	regoInstance := e.buildRegoInstance(query, input)

	resultSet, err := regoInstance.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("query eval: %w", err)
	}

	result := output.Result{
		Rule:   rule,
		Query:  query,
		Passed: len(resultSet) == 0,
	}

	return &result, nil
}

func (e Engine) querySkip(ctx context.Context, rule *output.Rule, input interface{}) (*output.Result, error) {
	query := fmt.Sprintf("data.%s.skip[_][_] == %q", rule.Namespace, rule.ID)
	regoInstance := e.buildRegoInstance(query, input)

	resultSet, err := regoInstance.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("skip query eval: %w", err)
	}

	result := output.Result{
		Rule:    rule,
		Query:   query,
		Skipped: len(resultSet) > 0,
	}

	return &result, nil
}

func (e Engine) buildRegoInstance(query string, input interface{}) *rego.Rego {
	return rego.New(
		rego.Query(query),
		rego.Input(input),
		rego.Compiler(e.compiler),
		rego.PrintHook(topdown.NewPrintHook(os.Stderr)),
	)
}

func allRegos(paths []string) (*loader.Result, error) {
	return loader.NewFileLoader().
		WithProcessAnnotation(true).
		Filtered(paths, func(_ string, info os.FileInfo, depth int) bool {
			return !info.IsDir() && !strings.HasSuffix(info.Name(), bundle.RegoExt)
		})
}

package provider

import (
	"encoding/json"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

type Provider interface {
	Builtins() []Builtin
	Schemas() []*Schema
}

// Builtin is a built-in function that can be used inside policies.
type Builtin interface {
	Func() *ast.Builtin
	Impl(rego.BuiltinContext, []*ast.Term) (*ast.Term, error)
}

type Schema struct {
	name       string
	raw        any
	jsonSchema *jsonschema.Schema
}

func NewSchema(name string, raw []byte) (*Schema, error) {
	schema := &Schema{name: name}

	if err := json.Unmarshal(raw, &schema.raw); err != nil {
		return nil, err
	}

	jsonSchema, err := jsonschema.CompileString(name, string(raw))
	if err != nil {
		return nil, err
	}
	schema.jsonSchema = jsonSchema

	return schema, nil
}

func (s Schema) Name() string {
	return s.name
}

func (s Schema) Validate(data any) error {
	return s.jsonSchema.Validate(data)
}

func (s Schema) Raw() any {
	return s.raw
}

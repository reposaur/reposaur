package provider

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/open-policy-agent/opa/rego"
)

var ErrNonDerivable = errors.New("data is non derivable")

type Namespace string

// DataDeriver is the interface that provides functions to derive a policy
// namespace and report properties from input data for a specific provider.
type DataDeriver interface {
	DeriveNamespace(map[string]any) (Namespace, error)
	DeriveProperties(Namespace, map[string]any) (map[string]any, error)
}

// Provider is the interface that provides the functions required to work with
// a specific provider, namely a function to register built-in functions in
// the policy engine and functions to derive required information from input data.
// See DataDeriver.
type Provider interface {
	DataDeriver

	Builtins() []Builtin
}

// Builtin is represents a built-in function in the policy engine. It specifies
// the function signature and the function implementation.
type Builtin struct {
	Func rego.Function
	Impl rego.BuiltinDyn
}

func DeriveNamespace(deriver DataDeriver, data any) (Namespace, error) {
	m, err := dataToMap(data)
	if err != nil {
		return "", err
	}

	return deriver.DeriveNamespace(m)
}

func DeriveProperties(deriver DataDeriver, namespace Namespace, data any) (map[string]any, error) {
	m, err := dataToMap(data)
	if err != nil {
		return nil, err
	}

	return deriver.DeriveProperties(namespace, m)
}

func dataToMap(data any) (map[string]any, error) {
	val := reflect.ValueOf(data)

	// We can never derive anything from a slice, only from maps or structures
	if val.Kind() == reflect.Slice {
		return nil, ErrNonDerivable
	}

	// If data is already a map we just convert it to map[string]any
	if val.Kind() == reflect.Map {
		m := make(map[string]any, val.Len())
		for _, k := range val.MapKeys() {
			if k.Kind() == reflect.String {
				m[k.String()] = val.MapIndex(k).Interface()
			}
		}
		return m, nil
	}

	if val.Kind() != reflect.Struct {
		return nil, ErrNonDerivable
	}

	// Here we have a structure, we rely on json package to convert it into
	// a map for us.
	rawData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var m map[string]any

	if err := json.Unmarshal(rawData, &m); err != nil {
		return nil, err
	}

	return m, nil
}

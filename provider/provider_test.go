package provider

import (
	"errors"
	"testing"
)

func TestMapDataToMap(t *testing.T) {
	dataMap := map[string]any{
		"foo": "bar",
		"bar": 10,
	}

	m, err := dataToMap(dataMap)
	if err != nil {
		t.Error(err)
	}

	if foo, ok := m["foo"]; !ok || foo != "bar" {
		t.Fail()
	}

	if bar, ok := m["bar"]; !ok || bar != 10 {
		t.Fail()
	}
}

func TestStructDataToMap(t *testing.T) {
	dataStruct := struct {
		Foo string `json:"foo"`
		Bar int    `json:"bar"`
	}{
		Foo: "bar",
		Bar: 10,
	}

	m, err := dataToMap(dataStruct)
	if err != nil {
		t.Error(err)
	}

	if foo, ok := m["foo"]; !ok || foo != "bar" {
		t.Fail()
	}

	// JSON numbers are stored as float64 when converting to interface{}
	// See https://pkg.go.dev/encoding/json#Unmarshal
	if bar, ok := m["bar"]; !ok || bar != 10.0 {
		t.Fail()
	}
}

func TestAnyDataToMap(t *testing.T) {
	_, err := dataToMap(10)
	if !errors.Is(err, ErrNonDerivable) {
		t.Errorf("expected error '%s' got '%s'", ErrNonDerivable, err)
	}

	_, err = dataToMap("foo")
	if !errors.Is(err, ErrNonDerivable) {
		t.Errorf("expected error '%s' got '%s'", ErrNonDerivable, err)
	}

	_, err = dataToMap([]string{"foo", "bar"})
	if !errors.Is(err, ErrNonDerivable) {
		t.Errorf("expected error '%s' got '%s'", ErrNonDerivable, err)
	}
}

package sdk

import (
	"errors"
	"reflect"
)

var ErrUnknownNamespace = errors.New("failed to detect namespace from data")

var namespaceToKeysMap = map[string][]string{
	"pull_request": {"base", "head"},
	"repository":   {"owner", "full_name"},
}

func DetectNamespace(data interface{}) (string, error) {
	val := reflect.ValueOf(data)

	if val.Kind() != reflect.Map {
		return "", ErrUnknownNamespace
	}

	for n, keys := range namespaceToKeysMap {
		var matches int

		for _, k := range keys {
			for _, dk := range val.MapKeys() {
				if k == dk.String() {
					matches++
				}
			}

			if matches == len(keys) {
				return n, nil
			}
		}
	}

	return "", ErrUnknownNamespace
}

package sdk

import (
	"errors"
	"reflect"
)

// ErrUnknownNamespace happens when DetectNamespace
// fails to detect a valid namespace.
var ErrUnknownNamespace = errors.New("failed to detect namespace from data")

var namespaceToKeysMap = map[string][]string{
	"issue":        {"reactions", "closed_by"},
	"organization": {"login", "members_url"},
	"pull_request": {"base", "head"},
	"repository":   {"owner", "full_name"},
	"user":         {"login", "hireable"},
}

// DetectNamespace will attempt to detect some data's
// namespace based on it's keys.
// 
// Returns an error of type ErrUnknownNamespace if data
// is not a map or if it can't detect a valid namespace.
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

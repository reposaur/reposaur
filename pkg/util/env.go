package util

import (
	"fmt"
	"os"
	"strconv"
)

// GetEnvOrDefault looks up every key provided in
// environment variables and returns the first value found.
// If no value is found, returns nil.
func GetEnv(keys ...string) *string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return &v
		}
	}

	return nil
}

func GetInt64Env(keys ...string) *int64 {
	val := GetEnv(keys...)

	if val == nil {
		return nil
	}

	intVal, err := strconv.ParseInt(*val, 10, 0)
	if err != nil {
		panic(fmt.Errorf("failed parsing env variable to int64: %w", err))
	}

	return &intVal
}

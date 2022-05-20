package cmdutil

import (
	"fmt"
	"os"
	"strconv"
)

// getEnv looks up every key in keys in the environment variables and returns
// the first value found. If no value is found, returns an empty string.
func getEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}

	return ""
}

// getInt64Env is similar to getEnv but converts the value to an int64. If no
// value is found returns the zero-value for int64.
func getInt64Env(keys ...string) int64 {
	val := getEnv(keys...)

	if val == "" {
		return 0
	}

	intVal, err := strconv.ParseInt(val, 10, 0)
	if err != nil {
		panic(fmt.Errorf("failed parsing env variable to int64: %w", err))
	}

	return intVal
}

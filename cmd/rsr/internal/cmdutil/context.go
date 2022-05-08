package cmdutil

import (
	"context"

	"github.com/rs/zerolog"
)

const LoggerKey = "RSR_LOGGER"

// ContextWithLogger returns a new context from ctx with
// a logger assigned to the "RSR_LOGGER" key.
func ContextWithLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

// LoggerFromContext returns an existing logger from ctx.
// If ctx doesn't have a logger a new one is created using buildContext.
func LoggerFromContext(ctx context.Context) zerolog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(zerolog.Logger); ok {
		return logger
	}

	return NewLogger(false)
}

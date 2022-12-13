package cmdutil

import (
	"os"

	"github.com/rs/zerolog"
)

// NewLogger returns a new logger that outputs to
// the standard error output. If verbose is true,
// the log level will be `debug`, otherwise will be info.
func NewLogger(verbose bool) zerolog.Logger {
	var (
		lvl = zerolog.InfoLevel
		cw  = zerolog.ConsoleWriter{Out: os.Stderr}
	)

	if verbose {
		lvl = zerolog.DebugLevel
	}

	return zerolog.New(cw).Level(lvl).With().Timestamp().Logger()
}

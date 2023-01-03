package cmdutil

import (
	"context"
	"golang.org/x/exp/slog"
	"io"
	"os"
)

// GetInputReader returns an io.ReadCloser. If filename
// is not empty, the file is opened and returned. Otherwise,
// returns a reader from standard input.
func GetInputReader(ctx context.Context, filename string) (r io.ReadCloser, err error) {
	var (
		logger = slog.FromContext(ctx)
		file   = os.Stdin
	)

	if filename == "" || filename == "-" {
		logger.Debug("using standard input as INPUT")
	} else {
		logger.Debug("using INPUT from file", "filename", filename)
		file, err = os.Open(filename)
		if err != nil {
			return nil, err
		}
	}

	return file, nil
}

// GetOutputWriter returns an io.WriteCloser. If filename
// is not empty, the file is opened and returned. Otherwise,
// returns a writer to standard output.
func GetOutputWriter(ctx context.Context, filename string) (w io.WriteCloser, err error) {
	var (
		logger = slog.FromContext(ctx)
		file   = os.Stdout
	)

	if filename == "" || filename == "-" {
		logger.Debug("using standard output as OUTPUT")
	} else {
		logger.Debug("using file as OUTPUT", "filename", filename)
		file, err = os.OpenFile(filename, os.O_WRONLY+os.O_CREATE+os.O_TRUNC, 0o666)
		if err != nil {
			return nil, err
		}
	}

	return file, nil
}

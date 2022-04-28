package cmdutil

import (
	"net/http"

	"github.com/rs/zerolog"
)

type Factory struct {
	Logger     zerolog.Logger
	HTTPClient *http.Client
}

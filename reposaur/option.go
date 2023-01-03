package reposaur

import (
	"github.com/go-git/go-billy/v5"
	"github.com/reposaur/reposaur/provider"
)

// Option is an option for the Reposaur.
type Option func(r *Reposaur)

// WithFilesystem specifies the billy.Filesystem used by the loader.FileLoader.
func WithFilesystem(fs billy.Filesystem) Option {
	return func(r *Reposaur) {
		r.fs = fs
	}
}

func WithProviders(providers ...provider.Provider) Option {
	return func(r *Reposaur) {
		for _, p := range providers {
			r.providers[providerName(p)] = p
		}
	}
}

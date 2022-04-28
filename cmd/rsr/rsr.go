package main

import (
	"os"

	"github.com/reposaur/reposaur/pkg/cmd/root"
	"github.com/reposaur/reposaur/pkg/cmdutil"
)

func main() {
	if err := root.NewCommand(&cmdutil.Factory{}).Execute(); err != nil {
		os.Exit(1)
	}
}

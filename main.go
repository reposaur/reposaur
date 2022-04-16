package main

import (
	"os"

	"github.com/reposaur/reposaur/cmd/reposaur"
)

func main() {
	if err := reposaur.NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

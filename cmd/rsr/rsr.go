package main

import (
	"fmt"
	"os"

	"github.com/reposaur/reposaur/cmd/rsr/internal/root"
)

func main() {
	if err := root.NewCmd().Execute(); err != nil {
		if _, err := fmt.Fprintln(os.Stderr, err); err != nil {
			panic(err)
		}
		os.Exit(1)
	}
}

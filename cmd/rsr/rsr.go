package main

import (
	"fmt"
	"os"

	"github.com/reposaur/reposaur/cmd/rsr/internal/root"
)

func main() {
	if err := root.NewCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

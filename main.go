package main

import (
	"context"
	"fmt"
	"os"

	"github.com/reposaur/reposaur/cmd/reposaur"
	"github.com/reposaur/reposaur/pkg/sdk"
)

var data = map[string]interface{}{
	"repository":   map[string]interface{}{},
	"pull_request": map[string]interface{}{},
}

func main() {
	if err := reposaur.NewCommand().Execute(); err != nil {
		os.Exit(1)
	}

	ctx := context.Background()

	rs, err := sdk.New(ctx, []string{})
	if err != nil {
		panic(err)
	}

	report, err := rs.Check(ctx, data)
	if err != nil {
		panic(err)
	}

	fmt.Println(report)
}

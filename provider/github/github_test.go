package github_test

import (
	"embed"
	"encoding/json"
	"github.com/reposaur/reposaur/provider/github"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

//go:embed testdata/*.json
var testdataFS embed.FS

func TestSchemas(t *testing.T) {
	t.Parallel()

	gh := &github.GitHub{}

	for _, schema := range gh.Schemas() {
		t.Run(schema.Name(), func(t *testing.T) {
			entries, err := testdataFS.ReadDir("testdata")
			require.NoError(t, err)

			for _, e := range entries {
				if !strings.HasPrefix(e.Name(), schema.Name()) {
					continue
				}

				t.Run(e.Name(), func(t *testing.T) {
					data, err := testdataFS.ReadFile("testdata/" + e.Name())
					require.NoError(t, err)

					var input any
					require.NoError(t, json.Unmarshal(data, &input))
					require.NoErrorf(t, schema.Validate(input), "testdata/%s isn't a valid %s", e.Name(), schema.Name())
				})
			}
		})
	}
}

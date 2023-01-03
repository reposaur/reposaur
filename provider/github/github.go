package github

import (
	"embed"
	"github.com/reposaur/reposaur/provider"
	"github.com/reposaur/reposaur/provider/github/client"
	"github.com/reposaur/reposaur/provider/github/internal/builtin"
	"strings"
)

//go:embed schemas/*.schema.json
var schemasFS embed.FS
var schemas []*provider.Schema

const schemaSuffix = ".schema.json"

type GitHub struct {
	client   *client.Client
	builtins []provider.Builtin
}

func New(c *client.Client) *GitHub {
	if c == nil {
		c = client.NewClient(nil)
	}

	return &GitHub{
		client: c,
		builtins: []provider.Builtin{
			&builtin.GraphQL{Client: c},
			&builtin.Request{Client: c},
		},
	}
}

func (gh GitHub) Builtins() []provider.Builtin {
	return gh.builtins
}

func (gh GitHub) Schemas() []*provider.Schema {
	return schemas
}

func init() {
	entries, err := schemasFS.ReadDir("schemas")
	if err != nil {
		panic(err)
	}

	for _, e := range entries {
		raw, err := schemasFS.ReadFile("schemas/" + e.Name())
		if err != nil {
			panic(err)
		}

		name := strings.TrimSuffix(e.Name(), schemaSuffix)

		schema, err := provider.NewSchema(name, raw)
		if err != nil {
			panic(err)
		}

		schemas = append(schemas, schema)
	}
}

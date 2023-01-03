package mock

import (
	"github.com/reposaur/reposaur/provider"
	"github.com/stretchr/testify/require"
	"io/fs"
	"path"
	"strings"
	"testing"
)

const schemaFileExt = ".schema.json"

type Mock struct {
	schemas []*provider.Schema
}

func New(t *testing.T, schemaFS fs.FS) *Mock {
	m := &Mock{}
	err := fs.WalkDir(schemaFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(p, schemaFileExt) {
			return nil
		}
		raw, err := fs.ReadFile(schemaFS, p)
		if err != nil {
			return err
		}
		schemaName := strings.TrimSuffix(path.Base(p), schemaFileExt)
		schema, err := provider.NewSchema(schemaName, raw)
		if err != nil {
			return err
		}
		m.schemas = append(m.schemas, schema)
		return nil
	})
	require.NoError(t, err)
	return m
}

func (m *Mock) Builtins() []provider.Builtin {
	return nil
}

func (m *Mock) Schemas() []*provider.Schema {
	return m.schemas
}

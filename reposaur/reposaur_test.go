package reposaur_test

import (
	"context"
	"embed"
	"encoding/json"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/reposaur/reposaur/provider/mock"
	"github.com/reposaur/reposaur/reposaur"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/fs"
	"path"
	"strings"
	"testing"
)

//go:embed testdata/data testdata/policies testdata/schemas
var testdataFS embed.FS

//nolint:funlen
func TestReposaur_LoadPolicies(t *testing.T) {
	t.Run("invalid schema fails", func(t *testing.T) {
		r := createTestInstance(t)
		require.ErrorContains(
			t,
			r.LoadPolicies("/invalid_schema.rego"),
			"undefined schema: schema.invalid.schema",
		)
	})

	t.Run("valid schema loads", func(t *testing.T) {
		r := createTestInstance(t)
		require.NoError(t, r.LoadPolicies("valid_schema.rego"))
	})

	t.Run("parse package metadata", func(t *testing.T) {
		r := createTestInstance(t)
		require.NoError(t, r.LoadPolicies("metadata.rego"))

		expected := reposaur.Metadata{
			Title:       "Example Package",
			Description: "Example Package Description",
			Authors: []reposaur.Author{
				{Name: "John Doe", Email: "john.doe@example.com"},
			},
			Organizations: []string{"Acme"},
		}

		require.EqualValues(t, expected, r.Policies()["testdata.metadata"].Metadata)
	})

	t.Run("parse rule metadata", func(t *testing.T) {
		r := createTestInstance(t)
		require.NoError(t, r.LoadPolicies("metadata.rego"))

		expected := reposaur.Metadata{
			Title:       "Example Rule",
			Description: "Example Rule Description",
			Authors: []reposaur.Author{
				{Name: "John Doe", Email: "john.doe@example.com"},
			},
			Organizations:    []string{"Acme"},
			Tags:             []string{"example", "guidelines"},
			SecuritySeverity: 3.5,
		}

		require.EqualValues(t, expected, r.Policies()["testdata.metadata"].Rules[0].Metadata)
	})

	t.Run("nested directories", func(t *testing.T) {
		r := createTestInstance(t, "./deep")
		require.Exactly(t, 1, len(r.Policies()))
	})

	t.Run("parse multiple rules", func(t *testing.T) {
		r := createTestInstance(t, "repo.rego", "pull.rego")
		lenRepoDiff := assert.Exactly(t, 1, len(r.Policies()["testdata.repo"].Rules))
		lenPullDiff := assert.Exactly(t, 1, len(r.Policies()["testdata.pull"].Rules))
		if !lenRepoDiff || !lenPullDiff {
			require.Fail(t, "errors parsing the policies")
		}
	})
}

func TestReposaur_Eval(t *testing.T) {
	data := readTestData(t)
	r := createTestInstance(t, "repo.rego", "pull.rego")

	t.Run("unrecognized input fails", func(t *testing.T) {
		_, err := r.Eval(context.Background(), data["invalid"])
		require.ErrorContains(t, err, "unknown schema for input")
	})

	t.Run("multiple rules", func(t *testing.T) {
		r, err := r.Eval(context.Background(), data["valid_repo"])
		require.NoError(t, err)
		s, err := r.SARIF()
		require.NoError(t, err)
		require.Equal(t, len(s.Runs[0].Results), 0)
	})
}

func createTestInstance(t *testing.T, loadPaths ...string) *reposaur.Reposaur {
	r, err := reposaur.New(
		reposaur.WithFilesystem(createTestFS(t)),
		reposaur.WithProviders(mock.New(t, testdataFS)),
	)
	require.NoError(t, err)

	if len(loadPaths) != 0 {
		err = r.LoadPolicies(loadPaths...)
		require.NoError(t, err)
	}

	return r
}

func readTestData(t *testing.T) map[string]any {
	root := "testdata/data"
	data := map[string]any{}

	entries, err := testdataFS.ReadDir(root)
	require.NoError(t, err)

	for _, e := range entries {
		raw, err := testdataFS.ReadFile(path.Join(root, e.Name()))
		require.NoError(t, err)

		var dataParsed any
		require.NoError(t, json.Unmarshal(raw, &dataParsed))

		name := strings.TrimSuffix(e.Name(), ".json")
		data[name] = dataParsed
	}

	return data
}

func createTestFS(t *testing.T) billy.Filesystem {
	root := "testdata/policies"
	mfs := memfs.New()
	err := fs.WalkDir(testdataFS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		newPath := strings.TrimPrefix(path, root)

		if d.IsDir() {
			if err := mfs.MkdirAll(path, 0777); err != nil {
				return err
			}
			return nil
		}

		raw, err := testdataFS.ReadFile(path)
		if err != nil {
			return err
		}

		f, err := mfs.Create(newPath)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = f.Write(raw)
		return err
	})
	require.NoError(t, err)

	return mfs
}

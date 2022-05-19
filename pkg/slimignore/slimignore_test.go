package slimignore

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestMatcher(t *testing.T) {
	type spec struct {
		ignoreFile    string
		fs            fs.FS
		expectIgnored map[string]bool
	}

	specs := []spec{
		{
			ignoreFile: ".slimignore",
			fs: fstest.MapFS{
				".slimignore": &fstest.MapFile{
					Data: []byte(`
						# Always ignore a anywhere
						a
						# Ignore the root ignore directory
						/ignore
						# Ignore files named "file1" in dir "really"
						/really/**/file1
					`),
				},
				"a":                             &fstest.MapFile{},
				"b":                             &fstest.MapFile{},
				"c":                             &fstest.MapFile{},
				"ignore/a":                      &fstest.MapFile{},
				"ignore/b":                      &fstest.MapFile{},
				"ignore/c":                      &fstest.MapFile{},
				"really/file1":                  &fstest.MapFile{},
				"really/file2":                  &fstest.MapFile{},
				"really/long/file1":             &fstest.MapFile{},
				"really/long/file2":             &fstest.MapFile{},
				"really/long/path/file1":        &fstest.MapFile{},
				"really/long/path/file2":        &fstest.MapFile{},
				"really/long/path/to/file1":     &fstest.MapFile{},
				"really/long/path/to/file2":     &fstest.MapFile{},
				"really/long/path/to/the/file1": &fstest.MapFile{},
				"really/long/path/to/the/file2": &fstest.MapFile{},
			},
			expectIgnored: map[string]bool{
				".":                             false,
				".slimignore":                   false,
				"a":                             true,
				"b":                             false,
				"c":                             false,
				"ignore":                        true,
				"ignore/a":                      true,
				"ignore/b":                      true,
				"ignore/c":                      true,
				"really":                        false,
				"really/file1":                  true,
				"really/file2":                  false,
				"really/long":                   false,
				"really/long/file1":             true,
				"really/long/file2":             false,
				"really/long/path":              false,
				"really/long/path/file1":        true,
				"really/long/path/file2":        false,
				"really/long/path/to":           false,
				"really/long/path/to/file1":     true,
				"really/long/path/to/file2":     false,
				"really/long/path/to/the":       false,
				"really/long/path/to/the/file1": true,
				"really/long/path/to/the/file2": false,
			},
		},
	}
	for _, s := range specs {
		require.NoError(t, fstest.TestFS(s.fs,
			".slimignore",
			"a",
			"b",
			"c",
			"ignore/a",
			"ignore/b",
			"ignore/c",
			"really/file1",
			"really/file2",
			"really/long/file1",
			"really/long/file2",
			"really/long/path/file1",
			"really/long/path/file2",
			"really/long/path/to/file1",
			"really/long/path/to/file2",
			"really/long/path/to/the/file1",
			"really/long/path/to/the/file2",
		),
		)

		m, err := newMatcher(s.fs, s.ignoreFile)
		require.NoError(t, err)
		err = fs.WalkDir(s.fs, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			expected, ok := s.expectIgnored[path]
			require.True(t, ok, "Found path %q in test FS, but did not find expectation for it", path)

			actual := m.Match(path, d.IsDir())
			require.Equal(t, expected, actual, "Expected ignore status for path %q to be %v, but got %v", path, expected, actual)
			return nil
		})
		require.NoError(t, err)
	}
}

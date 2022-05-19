package slimkeep

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate/version"
)

func TestMatcher(t *testing.T) {
	type spec struct {
		keepFile   string
		fs         fs.FS
		expectKept map[string]bool
	}

	const skData = `
#dslim.ver.info:file.ver=v1alpha1
#dslim.ver.info:dslim.ver=1.2.3

# Always keep a anywhere
a
# Keep the root keep directory
/keep
# Keep only files named "file1" in dir "really"
/really/**/file1
`

	specs := []spec{
		{
			keepFile: ".slimkeep",
			fs: fstest.MapFS{
				".slimkeep":                     &fstest.MapFile{Data: []byte(skData)},
				"a":                             &fstest.MapFile{},
				"b":                             &fstest.MapFile{},
				"c":                             &fstest.MapFile{},
				"keep/a":                        &fstest.MapFile{},
				"keep/b":                        &fstest.MapFile{},
				"keep/c":                        &fstest.MapFile{},
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
			expectKept: map[string]bool{
				".":                             false,
				".slimkeep":                     false,
				"a":                             true,
				"b":                             false,
				"c":                             false,
				"keep":                          true,
				"keep/a":                        true,
				"keep/b":                        true,
				"keep/c":                        true,
				"some/keep/a":                   true,
				"some/keep/b":                   false,
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
			".slimkeep",
			"a",
			"b",
			"c",
			"keep/a",
			"keep/b",
			"keep/c",
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

		m, err := loadMatcherFile(s.fs, s.keepFile)
		require.NoError(t, err)
		err = fs.WalkDir(s.fs, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			expected, ok := s.expectKept[path]
			require.True(t, ok, "Found path %q in test FS, but did not find expectation for it", path)

			actual := m.MatchInclude(pathSep+path, d.IsDir())
			require.Equal(t, expected, actual, "Expected ignore status for path %q to be %v, but got %v", path, expected, actual)
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, m.Version, version.Version{
			DockerSlimVersion: "1.2.3",
			FileVersion:       "v1alpha1",
		})
	}
}

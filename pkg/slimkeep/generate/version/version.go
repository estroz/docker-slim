package version

import (
	"fmt"
	"io"
	"strings"

	"github.com/docker-slim/docker-slim/pkg/version"
)

const (
	versionLinePrefix = "dslim.ver.info:"

	fileVerKey     = "file.ver"
	fileVerPrefix  = fileVerKey + "="
	dslimVerKey    = "dslim.ver"
	dslimVerPrefix = dslimVerKey + "="

	v1alpha1 = "v1alpha1"
)

type Version struct {
	DockerSlimVersion string
	FileVersion       string
}

func Write(w io.Writer) error {

	dslimVerInfo := version.Current()
	dslimVerSplit := strings.Split(dslimVerInfo, "|")
	if len(dslimVerSplit) != 5 {
		return fmt.Errorf("bad docker-slim version line, expected 5 parts separated by |, got: %q", dslimVerInfo)
	}
	// The semver component is the 3rd element.
	dslimSemver := dslimVerSplit[2]

	headerParts := [][]byte{
		{'#'},
		[]byte(versionLinePrefix),
		[]byte(fileVerPrefix),
		[]byte(v1alpha1),
		{'\n', '#'},
		[]byte(versionLinePrefix),
		[]byte(dslimVerPrefix),
		[]byte(dslimSemver),
		{'\n', '\n'},
	}

	for _, headerPart := range headerParts {
		if _, err := w.Write(headerPart); err != nil {
			return err
		}
	}

	return nil
}

const commentedPrefix = "#" + versionLinePrefix

func IsVersionLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), commentedPrefix)
}

func Parse(lines []string) (v Version, err error) {

	for _, line := range lines {
		line = strings.TrimPrefix(strings.TrimSpace(line), commentedPrefix)
		split := strings.Split(line, ",")
		for _, kv := range split {
			kvsplit := strings.SplitN(kv, "=", 2)
			if len(kvsplit) != 2 {
				return v, fmt.Errorf("malformed version line: %q", line)
			}
			key, val := kvsplit[0], kvsplit[1]
			switch key {
			case dslimVerKey:
				if v.DockerSlimVersion != "" {
					return v, fmt.Errorf("malformed version line: duplicate key %q", dslimVerKey)
				}
				v.DockerSlimVersion = val
			case fileVerKey:
				if v.FileVersion != "" {
					return v, fmt.Errorf("malformed version line: duplicate key %q", fileVerKey)
				}
				v.FileVersion = val
			default:
				return v, fmt.Errorf("malformed version line: unknown key in %q", kv)
			}
		}
	}

	return v, nil
}

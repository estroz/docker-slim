package slimkeep

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	log "github.com/sirupsen/logrus"

	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate/version"
)

const SlimKeepFile = ".slimkeep"

func NewDefaultMatcher() (*Matcher, bool, error) {
	info, err := os.Stat(SlimKeepFile)
	if (err == nil && !info.IsDir()) || errors.Is(err, os.ErrExist) {
		m, err := NewMatcher(SlimKeepFile)
		return m, true, err
	}
	return nil, false, nil
}

func NewMatcher(keepFile string) (*Matcher, error) {
	absFile, err := filepath.Abs(keepFile)
	if err != nil {
		return nil, err
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	relFile, err := filepath.Rel(wd, absFile)
	if err != nil {
		return nil, err
	}
	return loadMatcherFile(os.DirFS(wd), relFile)
}

type Matcher struct {
	Version version.Version

	patterns []gitignore.Pattern
	pats     []pat
}

type pat struct {
	Str     string   `json:"str"`
	Domains []string `json:"domains"`
}

func (m *Matcher) MatchInclude(path string, isDir bool) bool {
	// Gitignore is the inverse of slimkeep, since matches are kept.
	return m.match(path, isDir, gitignore.Exclude)
}

func (m *Matcher) MatchExclude(path string, isDir bool) bool {
	// Gitignore is the inverse of slimkeep, since matches are kept.
	return m.match(path, isDir, gitignore.Include)
}

const pathSep = "/"

func (m *Matcher) match(path string, isDir bool, cmpType gitignore.MatchResult) bool {
	if path[0] == '/' {
		path = path[1:]
	}
	paths := strings.Split(path, pathSep)
	n := len(m.patterns)
	for i := n - 1; i >= 0; i-- {
		if match := m.patterns[i].Match(paths, isDir); match > gitignore.NoMatch {
			return match == cmpType
		}
	}
	return false
}

func (m *Matcher) Add(p gitignore.Pattern, pattern, domain string) {
	m.pats = append(m.pats, pat{
		Str:     pattern,
		Domains: []string{domain},
	})
	m.patterns = append(m.patterns, p)
}

func (m *Matcher) AddPattern(pattern, domain string) {
	domains := []string{domain}
	m.pats = append(m.pats, pat{
		Str:     pattern,
		Domains: domains,
	})
	m.patterns = append(m.patterns, gitignore.ParsePattern(pattern, domains))
}

type matcherJSON struct {
	Pats []pat `json:"pats"`
}

func (m *Matcher) MarshalJSON() ([]byte, error) {
	return json.Marshal(matcherJSON{
		Pats: m.pats,
	})
}

func (m *Matcher) UnmarshalJSON(b []byte) error {
	var v matcherJSON
	if err := json.Unmarshal(b, &v); err != nil {
		return nil
	}

	m.pats = v.Pats
	m.patterns = make([]gitignore.Pattern, len(m.pats))
	for i, p := range m.pats {
		m.patterns[i] = gitignore.ParsePattern(p.Str, p.Domains)
	}

	return nil
}

func loadMatcherFile(root fs.FS, keepFile string) (*Matcher, error) {
	f, err := root.Open(keepFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m := &Matcher{}

	// There is no pattern domain since all matching happens at root.
	domains := []string{}
	var verLines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		switch {
		case version.IsVersionLine(line):
			verLines = append(verLines, line)
		case strings.HasPrefix(line, "#"), line == "":
			// Comment, ignore.
		default:
			m.patterns = append(m.patterns, gitignore.ParsePattern(line, domains))
			m.pats = append(m.pats, pat{Str: line, Domains: domains})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if m.Version, err = version.Parse(verLines); err != nil {
		log.Warnf("unable to parse slimkeep file version info: %v", err)
	}

	return m, nil
}

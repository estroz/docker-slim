package slimignore

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

const SlimIgnoreFile = ".slimignore"

type Matcher interface {
	Match(path string, isDir bool) bool
	Add(p gitignore.Pattern)
	AddPattern(pattern, domain string)
}

func NewDefaultMatcher() (Matcher, bool, error) {
	info, err := os.Stat(SlimIgnoreFile)
	if (err == nil && !info.IsDir()) || errors.Is(err, os.ErrExist) {
		m, err := NewMatcher(SlimIgnoreFile)
		return m, true, err
	}
	return nil, false, nil
}

func NewMatcher(ignoreFilePath string) (Matcher, error) {
	dir := ""
	if !filepath.IsAbs(ignoreFilePath) {
		dir = "."
	}
	return newMatcher(os.DirFS(dir), ignoreFilePath)
}

type matcher struct {
	m        gitignore.Matcher
	patterns []gitignore.Pattern
}

func (m *matcher) Match(path string, isDir bool) bool {
	return m.m.Match(strings.Split(path, string(filepath.Separator)), isDir)
}

func (m *matcher) Add(p gitignore.Pattern) {
	m.patterns = append(m.patterns, p)
	m.m = gitignore.NewMatcher(m.patterns)
}

func (m *matcher) AddPattern(pattern, domain string) {
	m.patterns = append(m.patterns, gitignore.ParsePattern(pattern, []string{domain}))
	m.m = gitignore.NewMatcher(m.patterns)
}

func newMatcher(root fs.FS, ignoreFilePath string) (Matcher, error) {
	patterns, err := loadPatterns(root, ignoreFilePath)
	if err != nil {
		return nil, err
	}
	return &matcher{
		m:        gitignore.NewMatcher(patterns),
		patterns: patterns,
	}, nil
}

func loadPatterns(root fs.FS, path string) ([]gitignore.Pattern, error) {
	domain := strings.Split(filepath.Dir(path), string(filepath.Separator))
	if len(domain) == 1 && domain[0] == "." {
		domain = []string{}
	}
	f, err := root.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	patterns := []gitignore.Pattern{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		patterns = append(patterns, gitignore.ParsePattern(line, domain))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return patterns, nil
}

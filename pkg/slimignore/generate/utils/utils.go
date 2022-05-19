package utils

import (
	"bytes"
	"strings"
)

type FileSection struct {
	bytes.Buffer
}

func (f *FileSection) WriteHeader(body string) {
	f.write(body)
}

func (f *FileSection) WriteComment(line string) {
	f.WriteString("# ")
	f.write(line)
}

func (f *FileSection) WriteIgnore(line string) {
	f.write(line)
}

func (f *FileSection) WriteKeep(line string) {
	f.WriteByte('!')
	f.write(line)
}

func (f *FileSection) WriteBlock(line string) {
	f.write(strings.TrimSpace(line))
}

func (f *FileSection) write(line string) {
	f.WriteString(line)
	f.WriteByte('\n')
}

func GlobAll(line string) string {
	const globAll = "/**"
	if line == "" {
		return line
	}
	if line[len(line)-2] == '/' {
		return line[:len(line)-1] + globAll
	}
	return line + globAll
}

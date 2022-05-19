package certs

import (
	"sort"

	"github.com/docker-slim/docker-slim/pkg/certdiscover"
	"github.com/docker-slim/docker-slim/pkg/slimignore/generate/utils"
)

func GenFileSection(keep bool) *utils.FileSection {
	f := &utils.FileSection{}

	var wf func(string)
	if keep {
		wf = f.WriteKeep
	} else {
		wf = f.WriteIgnore
	}

	// Certs.
	writeCerts(f, wf, keep)

	// Private keys.
	writePKs(f, wf, keep)

	return f
}

const (
	ignoreCertHeader = `
## System cert paths to ignore ##
`

	keepCertHeader = `
## System cert paths to keep ##
`
)

// Certs.
func writeCerts(f *utils.FileSection, wf func(string), keep bool) {
	if keep {
		f.WriteHeader(keepCertHeader)
	} else {
		f.WriteHeader(ignoreCertHeader)
	}

	certPathSet := map[string]struct{}{}
	for _, path := range certdiscover.CertDirList() {
		certPathSet[path] = struct{}{}
	}
	for _, path := range certdiscover.CACertDirList() {
		certPathSet[path] = struct{}{}
	}
	for _, path := range certdiscover.CertExtraDirList() {
		certPathSet[path] = struct{}{}
	}

	writeFileSet(wf, certPathSet)
}

const (
	ignorePKHeader = `
## System private key paths to ignore ##
`

	keepPKHeader = `
## System private key paths to keep ##
`
)

// Private keys.
func writePKs(f *utils.FileSection, wf func(string), keep bool) {
	if keep {
		f.WriteHeader(keepPKHeader)
	} else {
		f.WriteHeader(ignorePKHeader)
	}

	pkPathSet := map[string]struct{}{}
	for _, path := range certdiscover.CertPKDirList() {
		pkPathSet[path] = struct{}{}
	}
	for _, path := range certdiscover.CACertPKDirList() {
		pkPathSet[path] = struct{}{}
	}

	writeFileSet(wf, pkPathSet)
}

func writeFileSet(wf func(string), pathSet map[string]struct{}) {
	var paths []string
	for path := range pathSet {
		paths = append(paths, utils.GlobAll(path))
	}
	sort.Strings(paths)

	for _, path := range paths {
		wf(path)
	}
}

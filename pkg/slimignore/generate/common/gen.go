package common

import (
	"github.com/docker-slim/docker-slim/pkg/slimignore/generate/utils"
)

const header = `
## Common paths ##
`

func GenFileSection() *utils.FileSection {
	f := &utils.FileSection{}

	f.WriteHeader(header)
	writeContent(f)

	return f
}

const content = ``

func writeContent(f *utils.FileSection) {
	f.WriteBlock(content)
}

package common

import (
	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate/utils"
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

const content = `
# Unix system files, ex. those modifying $PATH vars.
/etc/profile
`

func writeContent(f *utils.FileSection) {
	f.WriteBlock(content)
}

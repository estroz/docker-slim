package nuxtjs

import (
	"fmt"

	"github.com/docker-slim/docker-slim/pkg/certdiscover"
	"github.com/docker-slim/docker-slim/pkg/slimignore/generate/appstack"
	"github.com/docker-slim/docker-slim/pkg/slimignore/generate/utils"
)

func init() {
	appstack.Register(func() appstack.AppStack { return &Stack{} })
}

const header = `
## NuxtJS paths ##
`

var _ appstack.AppStack = &Stack{}

type Stack struct{}

func (s *Stack) Name() string { return "nuxtjs" }

func (s *Stack) GenFileSection() *utils.FileSection {
	f := &utils.FileSection{}

	f.WriteHeader(header)
	writeConfigs(f)
	writeCerts(f)

	return f
}

const configsSection = `
# App source dir.
!/app
# NuxtJS config file.
!/**/nuxt.config.js
# Distribution dir.
!/**/dist
# Build dir.
!/**/.nuxt
# Static assets.
!/**/static
# JS modules.
!/**/node_modules
`

func writeConfigs(f *utils.FileSection) {
	f.WriteBlock(configsSection)
}

func writeCerts(f *utils.FileSection) {
	f.WriteByte('\n')
	f.WriteComment("Certs")
	f.WriteBlock(fmt.Sprintf("!/**/%s", certdiscover.AppCertPathSuffixNode))
	f.WriteBlock(fmt.Sprintf("!/%s", certdiscover.AppCertPathSuffixNode))
	f.WriteBlock(fmt.Sprintf("!/**/%s/**", certdiscover.AppCertPackageName))
	f.WriteBlock(fmt.Sprintf("!/%s/**", certdiscover.AppCertPackageName))
}

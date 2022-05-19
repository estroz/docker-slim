package nuxtjs

import (
	"fmt"

	"github.com/docker-slim/docker-slim/pkg/certdiscover"
	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate/appstack"
	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate/utils"
)

func init() {
	appstack.Register(func() appstack.AppStack { return &Stack{} })
}

const (
	name = "nuxtjs"

	header = `
## NuxtJS paths ##
`
)

var _ appstack.AppStack = &Stack{}

type Stack struct {
	gen func(string) *utils.FileSection
}

func (s *Stack) Name() string { return name }

func (s *Stack) GenFileSection() *utils.FileSection {
	if s.gen == nil {
		s.gen = defaultGen
	}

	f := s.gen("")

	return f
}

var defaultGen = func(prefix string) *utils.FileSection {
	f := &utils.FileSection{}

	f.WriteHeader(header)
	writeConfigs(f, prefix)
	writeCerts(f, prefix)

	return f
}

const configsSection = `
# NuxtJS config file.
%[1]s/app/nuxt.config.js
# Distribution dir.
%[1]s/app/dist/
# Build dir.
%[1]s/app/.nuxt/
# Static assets.
%[1]s/app/static/
# Content assets.
%[1]s/app/content/

# JS modules dir.
# Only uncomment this line if you are sure you want _all_ modules in your image.
# docker-slim will not remove untouched modules.
# You write a more specific pattern if you'd like.
#
# %[1]s/app/node_modules/
`

func writeConfigs(f *utils.FileSection, prefix string) {
	section := fmt.Sprintf(configsSection, prefix)
	f.WriteBlock(section)
}

func writeCerts(f *utils.FileSection, prefix string) {
	f.WriteComment("Certs")
	f.WriteKeep(fmt.Sprintf("%s/**/%s", prefix, certdiscover.AppCertPathSuffixNode))
	f.WriteKeep(fmt.Sprintf("%s/%s", prefix, certdiscover.AppCertPathSuffixNode))
	f.WriteKeep(fmt.Sprintf("%s/**/%s/**", prefix, certdiscover.AppCertPackageName))
	f.WriteKeep(fmt.Sprintf("%s/%s/**", prefix, certdiscover.AppCertPackageName))
}

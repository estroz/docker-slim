package ruby

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
## Ruby ignore statements ##
`

var _ appstack.AppStack = &Stack{}

type Stack struct{}

func (s *Stack) Name() string { return "ruby" }

func (s *Stack) GenFileSection() *utils.FileSection {
	f := &utils.FileSection{}

	f.WriteHeader(header)

	f.WriteComment("Certs")
	f.WriteKeep(fmt.Sprintf("/**/%s", certdiscover.AppCertPathSuffixRuby))
	f.WriteKeep(fmt.Sprintf("/%s", certdiscover.AppCertPathSuffixRuby))
	f.WriteKeep(fmt.Sprintf("/**/%s/**", certdiscover.AppCertPackageName))
	f.WriteKeep(fmt.Sprintf("/%s/**", certdiscover.AppCertPackageName))

	return f
}

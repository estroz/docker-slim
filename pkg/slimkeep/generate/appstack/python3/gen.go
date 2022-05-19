package python3

import (
	"fmt"

	"github.com/docker-slim/docker-slim/pkg/certdiscover"
	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate/appstack"
	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate/utils"
)

func init() {
	appstack.Register(func() appstack.AppStack { return &Stack{} })
}

const header = `
## Python3 ignore statements ##
`

type Stack struct{}

func (s *Stack) Name() string { return "python3" }

func (s *Stack) GenFileSection() *utils.FileSection {
	f := &utils.FileSection{}

	f.WriteHeader(header)

	f.WriteComment("Certs")
	f.WriteKeep(fmt.Sprintf("/**/%s", certdiscover.AppCertPathSuffixPython))
	f.WriteKeep(fmt.Sprintf("/%s", certdiscover.AppCertPathSuffixPython))
	f.WriteKeep(fmt.Sprintf("/**/%s/**", certdiscover.AppCertPackageName))
	f.WriteKeep(fmt.Sprintf("/%s/**", certdiscover.AppCertPackageName))

	return f
}

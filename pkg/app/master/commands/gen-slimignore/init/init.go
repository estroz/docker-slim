package init

import (
	genslimignore "github.com/docker-slim/docker-slim/pkg/app/master/commands/gen-slimignore"
)

func init() {
	genslimignore.RegisterCommand()
}

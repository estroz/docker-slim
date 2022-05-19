package init

import (
	genslimkeep "github.com/docker-slim/docker-slim/pkg/app/master/commands/gen-slimkeep"
)

func init() {
	genslimkeep.RegisterCommand()
}

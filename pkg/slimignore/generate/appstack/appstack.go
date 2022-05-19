package appstack

import (
	"sort"

	"github.com/docker-slim/docker-slim/pkg/slimignore/generate/utils"
)

type AppStack interface {
	Name() string
	GenFileSection() *utils.FileSection
}

type AppStacks []AppStack

func (as AppStacks) Sort() {
	sort.Slice(as, func(i, j int) bool { return as[i].Name() < as[j].Name() })
}

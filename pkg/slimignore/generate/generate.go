package generate

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker-slim/docker-slim/pkg/slimignore/generate/appstack"
	"github.com/docker-slim/docker-slim/pkg/slimignore/generate/certs"
	"github.com/docker-slim/docker-slim/pkg/slimignore/generate/common"
	"github.com/docker-slim/docker-slim/pkg/slimignore/generate/utils"
)

type Generator struct {
	Stacks    []appstack.AppStack
	KeepCerts bool
}

func (g *Generator) RunFile(ctx context.Context, filePath string) error {
	if dir := filepath.Dir(filePath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return fmt.Errorf("create .slimignore dir: %v", err)
		}
	}

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("create .slimignore file: %v", err)
	}
	defer f.Close()

	return g.Run(ctx, f)
}

const header = `#############################################################################
### This is a .slimignore file, consumed by 'docker-slim build' to prune  ###
### unneeded files from your images. Feel free to tailor the contents     ###
### of this file to your application's specific needs!                    ###
#############################################################################`

func (g *Generator) Run(ctx context.Context, out io.Writer) error {

	headerSection := &utils.FileSection{}
	headerSection.WriteHeader(header)
	sections := []*utils.FileSection{headerSection}
	// All user-specified app stacks.
	sections = append(sections, getSectionsForAppStacks(g.Stacks)...)
	// Common stuff.
	sections = append(sections, common.GenFileSection())
	// System certs and private keys.
	sections = append(sections, certs.GenFileSection(g.KeepCerts))

	for _, section := range sections {
		b := section.Bytes()
		if _, err := out.Write(b); err != nil {
			return fmt.Errorf("write .slimignore section: %v", err)
		}
	}

	return nil
}

func MakeAppStacks(stackNames []string) (stacks appstack.AppStacks, err error) {

	stackFuncMap := appstack.GetAll()

	var unavailableStacks []string
	for _, stackName := range stackNames {
		newStack, hasStack := stackFuncMap[stackName]
		if !hasStack {
			unavailableStacks = append(unavailableStacks, stackName)
			continue
		}
		stacks = append(stacks, newStack())
	}

	if len(unavailableStacks) != 0 {
		return nil, fmt.Errorf("these stacks are not implemented yet: %s", strings.Join(unavailableStacks, ", "))
	}

	return stacks, nil
}

func ListAllAppStacks() map[string]appstack.AppStack {
	stackFuncMap := appstack.GetAll()

	stackMap := make(map[string]appstack.AppStack, len(stackFuncMap))
	for stackName, newStack := range stackFuncMap {
		stackMap[stackName] = newStack()
	}

	return stackMap
}

func getSectionsForAppStacks(stacks []appstack.AppStack) []*utils.FileSection {
	copyStacks := make(appstack.AppStacks, len(stacks))
	copy(copyStacks, stacks)
	copyStacks.Sort()

	sections := make([]*utils.FileSection, len(copyStacks))
	for i, stack := range copyStacks {
		sections[i] = stack.GenFileSection()
	}

	return sections
}

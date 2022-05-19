package generate

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate/appstack"
	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate/certs"
	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate/common"
	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate/utils"
	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate/version"
)

type Generator struct {
	Stacks    []appstack.AppStack
	KeepCerts bool
}

func (g *Generator) RunFile(ctx context.Context, filePath string) error {
	if dir := filepath.Dir(filePath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return fmt.Errorf("create .slimkeep dir: %v", err)
		}
	}

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("create .slimkeep file: %v", err)
	}
	defer f.Close()

	return g.Run(ctx, f)
}

// TODO(estroz): this could be more explanatory as to how this works,
// and what *not* to add, ex the whole /app dir for NuxtJS apps.
const header = `###############################################################################
### This is a .slimkeep file, which tells 'docker-slim build' to explicitly ###
### keey necessary files in your images. Think of this file as the inverse  ###
### of a .gitignore, complemented by docker-slim's file access tracing      ###
### capabilities. Feel free to tailor the contents of this file to your     ###
### application's specific needs!                                           ###
###############################################################################`

func (g *Generator) Run(ctx context.Context, out io.Writer) error {

	verSection := &utils.FileSection{}
	if err := version.Write(verSection); err != nil {
		return fmt.Errorf("write version: %v", err)
	}

	headerSection := &utils.FileSection{}
	headerSection.WriteHeader(header)

	sections := []*utils.FileSection{verSection, headerSection}
	// All user-specified app stacks.
	sections = append(sections, getSectionsForAppStacks(g.Stacks)...)
	// Common stuff.
	sections = append(sections, common.GenFileSection())
	// System certs and private keys.
	sections = append(sections, certs.GenFileSection(g.KeepCerts))

	for _, section := range sections {
		b := section.Bytes()
		if _, err := out.Write(b); err != nil {
			return fmt.Errorf("write .slimkeep section: %v", err)
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

func GetAppStack(stackName string) (stack appstack.AppStack, err error) {
	stackFuncMap := appstack.GetAll()

	newStack, hasStack := stackFuncMap[stackName]
	if !hasStack {
		return nil, fmt.Errorf("this stack is not implemented yet: %s", stackName)
	}

	return newStack(), nil
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

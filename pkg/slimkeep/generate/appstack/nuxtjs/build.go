package nuxtjs

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate/appstack"
	"github.com/docker-slim/docker-slim/pkg/slimkeep/generate/utils"
)

var _ appstack.Builder = &Stack{}

func (s *Stack) BuilderID() string { return name }

func (s *Stack) Build(root fs.FS, prefix string) error {
	f := &utils.FileSection{}

	f.WriteHeader(header)

	foundConfig := false

	if err := fs.WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil {
			return err
		}

		nuxtConfig, ok, err := loadNuxtConfig(path)
		if err != nil || !ok {
			return err
		}

		foundConfig = true

		nuxtAppDir := filepath.Dir(path)
		nuxtAppDirPrefix := utils.DirLine(prefix + nuxtAppDir)
		if !filepath.IsAbs(nuxtAppDirPrefix) {
			nuxtAppDirPrefix = "/" + nuxtAppDirPrefix
		}
		f.WriteComment("App source dir.")
		f.WriteKeep(nuxtAppDirPrefix)

		staticDir := nuxtAppDirPrefix + nuxtStaticDir
		f.WriteComment("Static assets.")
		f.WriteKeep(utils.DirLine(staticDir))

		if nuxtConfig.Build != "" {
			buildDir := nuxtConfig.Build
			if !filepath.IsAbs(buildDir) {
				buildDir = nuxtAppDirPrefix + buildDir
			}

			f.WriteComment("Build dir.")
			f.WriteKeep(utils.DirLine(buildDir))
		}

		if nuxtConfig.Dist != "" {
			distDir := nuxtConfig.Dist
			if !filepath.IsAbs(distDir) {
				distDir = nuxtAppDirPrefix + distDir
			}

			f.WriteComment("Distribution dir.")
			f.WriteKeep(utils.DirLine(distDir))
		}

		f.WriteLine(`
# Only uncomment this line if you are sure you want _all_ modules in your image.
# docker-slim will not remove untouched modules.
# You write a more specific pattern if you'd like.
#`)
		modulesDir := nuxtAppDirPrefix + nodeModulesDir
		f.WriteComment(utils.DirLine(modulesDir))
		f.WriteByte('\n')

		return utils.ErrReturnEarly
	}); err != nil && !utils.IsReturnEarlyErr(err) {
		return err
	}

	if !foundConfig {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		fmt.Printf("Found no NuxtJS config in %s, using defaults with prefix %q\n",
			filepath.Join(wd, nuxtConfigFile), prefix)

		s.gen = func(_ string) *utils.FileSection {
			return defaultGen(prefix)
		}

		return nil
	}

	s.gen = func(string) *utils.FileSection {
		writeCerts(f, prefix)
		return f
	}

	return nil
}

const (
	nuxtConfigFile      = "nuxt.config.js"
	nuxtDefaultDistDir  = "dist"
	nuxtDefaultBuildDir = ".nuxt"
	nuxtStaticDir       = "static"
	nodeModulesDir      = "node_modules"
)

type nuxtConfig struct {
	Build string
	Dist  string
}

func isNuxtConfigFile(filePath string) bool {
	path := filepath.Base(filePath)
	return path == nuxtConfigFile
}

func loadNuxtConfig(path string) (*nuxtConfig, bool, error) {
	if !isNuxtConfigFile(path) {
		return nil, false, nil
	}

	_, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, false, err
	}

	//TODO: read the file and verify that it's a real nuxt config file

	nuxt := nuxtConfig{
		Build: nuxtDefaultBuildDir,
		Dist:  fmt.Sprintf("%s/%s", nuxtDefaultBuildDir, nuxtDefaultDistDir),
	}

	return &nuxt, true, nil
}

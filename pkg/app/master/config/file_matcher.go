package config

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/docker-slim/docker-slim/pkg/slimignore"
	"github.com/docker-slim/docker-slim/pkg/util/fsutil"
)

type FileMatcherConfig struct {
	Matcher slimignore.Matcher `json:"-"`

	fileMatcherConfigLegacy `json:",inline"`
}

func NewFileMatcherConfig(ctx *cli.Context, ignorePath string) (cfg *FileMatcherConfig, err error) {
	cfg = &FileMatcherConfig{}

	if ignorePath != "" {
		cfg.Matcher, err = slimignore.NewMatcher(ignorePath)
	} else {
		cfg.Matcher, _, err = slimignore.NewDefaultMatcher()
	}
	if err != nil {
		return nil, fmt.Errorf(".slimignore matcher: %v", err)
	}

	if cfg.Matcher != nil {
		cfg.Matcher.AddPattern("/opt/dockerslim", "")
		cfg.Matcher.AddPattern("/opt/dockerslim/**", "")
	}

	return cfg, nil
}

type fileMatcherConfigLegacy struct {
	// Global
	ExcludePatterns map[string]*fsutil.AccessInfo
	ExcludeMounts   bool

	// Build
	PreservePaths                map[string]*fsutil.AccessInfo
	IncludePaths                 map[string]*fsutil.AccessInfo
	PathPerms                    map[string]*fsutil.AccessInfo
	IncludeBins                  map[string]*fsutil.AccessInfo
	IncludeExes                  map[string]*fsutil.AccessInfo
	IncludeShell                 bool
	IncludeCertAll               bool
	IncludeCertBundles           bool
	IncludeCertDirs              bool
	IncludeCertPKAll             bool
	IncludeCertPKDirs            bool
	IncludeNew                   bool
	KeepTmpArtifacts             bool
	IncludeAppNuxtDir            bool
	IncludeAppNuxtBuildDir       bool
	IncludeAppNuxtDistDir        bool
	IncludeAppNuxtStaticDir      bool
	IncludeAppNuxtNodeModulesDir bool
	IncludeAppNextDir            bool
	IncludeAppNextBuildDir       bool
	IncludeAppNextDistDir        bool
	IncludeAppNextStaticDir      bool
	IncludeAppNextNodeModulesDir bool
	IncludeNodePackages          []string
}

func (c *FileMatcherConfig) MarshalJSON() ([]byte, error) {
	v := fileMatcherConfigLegacyJSON{
		Excludes:                     pathMapKeys(c.ExcludePatterns),
		Preserves:                    c.PreservePaths,
		Includes:                     c.IncludePaths,
		Perms:                        c.PathPerms,
		IncludeBins:                  pathMapKeys(c.IncludeBins),
		IncludeExes:                  pathMapKeys(c.IncludeExes),
		IncludeShell:                 c.IncludeShell,
		IncludeCertAll:               c.IncludeCertAll,
		IncludeCertBundles:           c.IncludeCertBundles,
		IncludeCertDirs:              c.IncludeCertDirs,
		IncludeCertPKAll:             c.IncludeCertPKAll,
		IncludeCertPKDirs:            c.IncludeCertPKDirs,
		IncludeNew:                   c.IncludeNew,
		IncludeAppNuxtDir:            c.IncludeAppNuxtDir,
		IncludeAppNuxtBuildDir:       c.IncludeAppNuxtBuildDir,
		IncludeAppNuxtDistDir:        c.IncludeAppNuxtDistDir,
		IncludeAppNuxtStaticDir:      c.IncludeAppNuxtStaticDir,
		IncludeAppNuxtNodeModulesDir: c.IncludeAppNuxtNodeModulesDir,
		IncludeAppNextDir:            c.IncludeAppNextDir,
		IncludeAppNextBuildDir:       c.IncludeAppNextBuildDir,
		IncludeAppNextDistDir:        c.IncludeAppNextDistDir,
		IncludeAppNextStaticDir:      c.IncludeAppNextStaticDir,
		IncludeAppNextNodeModulesDir: c.IncludeAppNextNodeModulesDir,
		IncludeNodePackages:          c.IncludeNodePackages,
	}

	return json.Marshal(v)
}

type fileMatcherConfigLegacyJSON struct {
	Excludes                     []string                      `json:"excludes,omitempty"`
	Preserves                    map[string]*fsutil.AccessInfo `json:"preserves,omitempty"`
	Includes                     map[string]*fsutil.AccessInfo `json:"includes,omitempty"`
	Perms                        map[string]*fsutil.AccessInfo `json:"perms,omitempty"`
	IncludeBins                  []string                      `json:"include_bins,omitempty"`
	IncludeExes                  []string                      `json:"include_exes,omitempty"`
	IncludeShell                 bool                          `json:"include_shell,omitempty"`
	IncludeCertAll               bool                          `json:"include_cert_all,omitempty"`
	IncludeCertBundles           bool                          `json:"include_cert_bundles,omitempty"`
	IncludeCertDirs              bool                          `json:"include_cert_dirs,omitempty"`
	IncludeCertPKAll             bool                          `json:"include_cert_pk_all,omitempty"`
	IncludeCertPKDirs            bool                          `json:"include_cert_pk_dirs,omitempty"`
	IncludeNew                   bool                          `json:"include_new,omitempty"`
	IncludeAppNuxtDir            bool                          `json:"include_app_nuxt_dir,omitempty"`
	IncludeAppNuxtBuildDir       bool                          `json:"include_app_nuxt_build,omitempty"`
	IncludeAppNuxtDistDir        bool                          `json:"include_app_nuxt_dist,omitempty"`
	IncludeAppNuxtStaticDir      bool                          `json:"include_app_nuxt_static,omitempty"`
	IncludeAppNuxtNodeModulesDir bool                          `json:"include_app_nuxt_nm,omitempty"`
	IncludeAppNextDir            bool                          `json:"include_app_next_dir,omitempty"`
	IncludeAppNextBuildDir       bool                          `json:"include_app_next_build,omitempty"`
	IncludeAppNextDistDir        bool                          `json:"include_app_next_dist,omitempty"`
	IncludeAppNextStaticDir      bool                          `json:"include_app_next_static,omitempty"`
	IncludeAppNextNodeModulesDir bool                          `json:"include_app_next_nm,omitempty"`
	IncludeNodePackages          []string                      `json:"include_node_packages,omitempty"`
}

func pathMapKeys(m map[string]*fsutil.AccessInfo) []string {
	if len(m) == 0 {
		return nil
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

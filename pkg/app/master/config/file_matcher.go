package config

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/docker-slim/docker-slim/pkg/slimkeep"
	"github.com/docker-slim/docker-slim/pkg/util/fsutil"
)

type FileMatcherConfig struct {
	Matcher *slimkeep.Matcher `json:"matcher,omitempty"`

	fileMatcherConfigLegacy `json:",inline"`
}

func NewFileMatcherConfig(ctx *cli.Context, ignorePath string) (cfg *FileMatcherConfig, err error) {
	cfg = &FileMatcherConfig{}

	if ignorePath != "" {
		cfg.Matcher, err = slimkeep.NewMatcher(ignorePath)
	} else {
		cfg.Matcher, _, err = slimkeep.NewDefaultMatcher()
	}
	if err != nil {
		return nil, fmt.Errorf(".slimkeep matcher: %v", err)
	}

	if cfg.Matcher != nil {
		cfg.Matcher.AddPattern("!/opt/dockerslim", "")
		cfg.Matcher.AddPattern("!/opt/dockerslim/**", "")
	}

	return cfg, nil
}

type fileMatcherConfigLegacy struct {
	// Global
	ExcludePatterns map[string]*fsutil.AccessInfo `json:"exclude_patterns,omitempty"`
	ExcludeMounts   bool                          `json:"exclude_mounts,omitempty"`

	// Build
	PreservePaths                map[string]*fsutil.AccessInfo `json:"preserve_paths,omitempty"`
	IncludePaths                 map[string]*fsutil.AccessInfo `json:"include_paths,omitempty"`
	PathPerms                    map[string]*fsutil.AccessInfo `json:"path_perms,omitempty"`
	IncludeBins                  map[string]*fsutil.AccessInfo `json:"include_bins,omitempty"`
	IncludeExes                  map[string]*fsutil.AccessInfo `json:"include_exes,omitempty"`
	IncludeShell                 bool                          `json:"include_shell,omitempty"`
	IncludeCertAll               bool                          `json:"include_cert_all,omitempty"`
	IncludeCertBundles           bool                          `json:"include_cert_bundles,omitempty"`
	IncludeCertDirs              bool                          `json:"include_cert_dirs,omitempty"`
	IncludeCertPKAll             bool                          `json:"include_cert_pk_all,omitempty"`
	IncludeCertPKDirs            bool                          `json:"include_cert_pk_dirs,omitempty"`
	IncludeNew                   bool                          `json:"include_new,omitempty"`
	KeepTmpArtifacts             bool                          `json:"keep_tmp_artifacts,omitempty"`
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

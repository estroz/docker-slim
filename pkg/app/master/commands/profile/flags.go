package profile

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/docker-slim/docker-slim/pkg/app/master/commands"
	"github.com/docker-slim/docker-slim/pkg/app/master/config"
)

func GetFileMatcherConfig(ctx *cli.Context, volumeMounts map[string]config.VolumeMount) (cfg *config.FileMatcherConfig, err error) {
	cfg, err = config.NewFileMatcherConfig(ctx, ctx.String(commands.FlagSlimignore))
	if err != nil {
		return nil, err
	}
	if cfg.Matcher != nil {
		badFlags := []string{}
		for f := range legacyFileMatcherConfigFlags {
			if ctx.IsSet(f) {
				badFlags = append(badFlags, f)
			}
		}
		if len(badFlags) != 0 {
			return nil, fmt.Errorf("cannot set legacy flags with .slimignore present: %s", strings.Join(badFlags, ", "))
		}
		return cfg, nil
	}

	return newFileMatcherConfigLegacy(ctx, volumeMounts)
}

func newFileMatcherConfigLegacy(ctx *cli.Context, volumeMounts map[string]config.VolumeMount) (cfg *config.FileMatcherConfig, err error) {
	cfg = &config.FileMatcherConfig{}

	cfg.ExcludePatterns = commands.ParsePaths(ctx.StringSlice(commands.FlagExcludePattern))
	cfg.ExcludeMounts = ctx.Bool(commands.FlagExcludeMounts)
	if cfg.ExcludeMounts {
		for mpath := range volumeMounts {
			cfg.ExcludePatterns[mpath] = nil
			mpattern := fmt.Sprintf("%s/**", mpath)
			cfg.ExcludePatterns[mpattern] = nil
		}
	}

	// cfg.PreservePaths, err = parsePathsLegacy(ctx, FlagPreservePath, FlagPreservePathFile)
	// if err != nil {
	// 	return nil, err
	// }
	// cfg.IncludePaths, err = parsePathsLegacy(ctx, FlagIncludePath, FlagIncludePathFile)
	// if err != nil {
	// 	return nil, err
	// }
	// cfg.PathPerms, err = parsePathsLegacy(ctx, FlagPathPerms, FlagPathPermsFile)
	// if err != nil {
	// 	return nil, err
	// }
	// cfg.IncludeBins, err = parsePathsLegacy(ctx, FlagIncludeBin, FlagIncludeBinFile)
	// if err != nil {
	// 	return nil, err
	// }
	// cfg.IncludeExes, err = parsePathsLegacy(ctx, FlagIncludeExe, FlagIncludeExeFile)
	// if err != nil {
	// 	return nil, err
	// }

	// cfg.IncludeShell = ctx.Bool(FlagIncludeShell)

	// cfg.IncludeCertAll = ctx.Bool(FlagIncludeCertAll)
	// cfg.IncludeCertBundles = ctx.Bool(FlagIncludeCertBundles)
	// cfg.IncludeCertDirs = ctx.Bool(FlagIncludeCertDirs)
	// cfg.IncludeCertPKAll = ctx.Bool(FlagIncludeCertPKAll)
	// cfg.IncludeCertPKDirs = ctx.Bool(FlagIncludeCertPKDirs)

	// cfg.IncludeNew = ctx.Bool(FlagIncludeNew)

	// cfg.KeepTmpArtifacts = ctx.Bool(FlagKeepTmpArtifacts)

	// cfg.IncludeAppNuxtDir = ctx.Bool(FlagIncludeAppNuxtDir)
	// cfg.IncludeAppNuxtBuildDir = ctx.Bool(FlagIncludeAppNuxtBuildDir)
	// cfg.IncludeAppNuxtDistDir = ctx.Bool(FlagIncludeAppNuxtDistDir)
	// cfg.IncludeAppNuxtStaticDir = ctx.Bool(FlagIncludeAppNuxtStaticDir)
	// cfg.IncludeAppNuxtNodeModulesDir = ctx.Bool(FlagIncludeAppNuxtNodeModulesDir)

	// cfg.IncludeAppNextDir = ctx.Bool(FlagIncludeAppNextDir)
	// cfg.IncludeAppNextBuildDir = ctx.Bool(FlagIncludeAppNextBuildDir)
	// cfg.IncludeAppNextDistDir = ctx.Bool(FlagIncludeAppNextDistDir)
	// cfg.IncludeAppNextStaticDir = ctx.Bool(FlagIncludeAppNextStaticDir)
	// cfg.IncludeAppNextNodeModulesDir = ctx.Bool(FlagIncludeAppNextNodeModulesDir)

	// cfg.IncludeNodePackages = ctx.StringSlice(FlagIncludeNodePackage)

	return cfg, nil
}

// legacyFileMatcherConfigFlags holds all flags that can be captured by a .slimignore.
var legacyFileMatcherConfigFlags = map[string]struct{}{
	commands.FlagExcludePattern: {},
	commands.FlagExcludeMounts:  {},
	// FlagPreservePath:                 {},
	// FlagPreservePathFile:             {},
	// FlagIncludePath:                  {},
	// FlagIncludePathFile:              {},
	// FlagPathPerms:                    {},
	// FlagPathPermsFile:                {},
	// FlagIncludeBin:                   {},
	// FlagIncludeBinFile:               {},
	// FlagIncludeExe:                   {},
	// FlagIncludeExeFile:               {},
	// FlagIncludeShell:                 {},
	// FlagIncludeCertAll:               {},
	// FlagIncludeCertBundles:           {},
	// FlagIncludeCertDirs:              {},
	// FlagIncludeCertPKAll:             {},
	// FlagIncludeCertPKDirs:            {},
	// FlagIncludeNew:                   {},
	// FlagKeepTmpArtifacts:             {},
	// FlagIncludeAppNuxtDir:            {},
	// FlagIncludeAppNuxtBuildDir:       {},
	// FlagIncludeAppNuxtDistDir:        {},
	// FlagIncludeAppNuxtStaticDir:      {},
	// FlagIncludeAppNuxtNodeModulesDir: {},
	// FlagIncludeAppNextDir:            {},
	// FlagIncludeAppNextBuildDir:       {},
	// FlagIncludeAppNextDistDir:        {},
	// FlagIncludeAppNextStaticDir:      {},
	// FlagIncludeAppNextNodeModulesDir: {},
	// FlagIncludeNodePackage:           {},
}

// func parsePathsLegacy(ctx *cli.Context, pathFlag, pathFileFlag string) (map[string]*fsutil.AccessInfo, error) {
// 	paths := commands.ParsePaths(ctx.StringSlice(pathFlag))
// 	morePaths, err := commands.ParsePathsFile(ctx.String(pathFileFlag))
// 	if err != nil {
// 		flagSplit := strings.Split(pathFileFlag, "-")
// 		return nil, fmt.Errorf("parse %s: %v", strings.Join(flagSplit, " "), err)
// 	}
// 	for k, v := range morePaths {
// 		paths[k] = v
// 	}

// 	return paths, nil
// }

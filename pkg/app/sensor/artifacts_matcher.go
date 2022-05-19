package app

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/armon/go-radix"
	log "github.com/sirupsen/logrus"

	"github.com/docker-slim/docker-slim/pkg/app/sensor/detectors/binfile"
	"github.com/docker-slim/docker-slim/pkg/app/sensor/inspectors/sodeps"
	"github.com/docker-slim/docker-slim/pkg/report"
	"github.com/docker-slim/docker-slim/pkg/slimkeep"
	"github.com/docker-slim/docker-slim/pkg/util/fsutil"
)

func (p *artifactStore) saveArtifactsMatcher(root fs.FS) {
	oldmask := syscall.Umask(0)
	defer syscall.Umask(oldmask)

	matcher := p.cmd.FileMatcherConfig.Matcher

	// TODO(estroz): thread these through to artifactStore.
	newPerms := map[string]*fsutil.AccessInfo{}
	extraDirs := map[string]struct{}{}
	symlinkFailed := map[string]*report.ArtifactProps{}

	log.Debugf("saveArtifacts - checking %d symlinks", len(p.linkMap))
	//copyLinks:
	//NOTE: MUST copy the links FIRST, so the dir symlinks get created before their files are copied
	symlinkMap := radix.New()
	for linkName, linkProps := range p.linkMap {
		symlinkMap.Insert(linkName, linkProps)
	}

	symlinkMap.Walk(func(linkPath string, val interface{}) bool {
		linkProps, ok := val.(*report.ArtifactProps)
		if !ok {
			log.Warnf("saveArtifacts.symlinkWalk: could not convert data (%s)", linkPath)
			return false
		}

		if continueOn(linkPath, linkProps, matcher) {
			return false
		}

		if tryLater := p.trySymlink(linkPath, linkProps); tryLater {
			//save it and try again later
			symlinkFailed[linkPath] = linkProps
		}

		return false
	})
	// Try symlinking once more.
	for linkName, linkProps := range symlinkFailed {
		p.trySymlink(linkName, linkProps)
	}

	// TODO(estroz): parallelize copies, since some of these copyDst calls will execute a command.

	for srcPath, srcProps := range p.fileMap {
		//filter out pid files (todo: have a flag to enable/disable these capabilities)
		if isKnownPidFilePath(srcPath) || hasPidFileSuffix(srcPath) {
			log.Debugf("saveArtifacts - copy files - adding dir of pid file %s", srcPath)
			extraDirs[fsutil.FileDir(srcPath)] = struct{}{}
			continue
		}

		if continueOn(srcPath, srcProps, matcher) {
			continue
		}

		p.copyDst(srcPath, srcProps, newPerms)
	}

	for srcPath, srcProps := range p.saFileMap {
		if continueOn(srcPath, srcProps, matcher) {
			continue
		}

		p.copyDst(srcPath, srcProps, newPerms)
	}

	const (
		passwdFile = "/etc/passwd"
		tmpDir     = "/tmp"
		runDir     = "/run"
	)

	if (matcher.MatchInclude(passwdFile, false) && !matcher.MatchExclude(passwdFile, false)) || p.cmd.AppUser != "" {
		//always copy the '/etc/passwd' file when we have a user
		//later: do it only when AppUser is a name (not UID)
		passwdFileTargetPath := p.toDstPath(passwdFile)
		if _, err := os.Stat(passwdFile); err == nil {
			if err := fsutil.CopyRegularFile(p.cmd.KeepPerms, passwdFile, passwdFileTargetPath, true); err != nil {
				log.Warnf("sensor: monitor - error copying user info file => %s", err)
			}
		} else if errors.Is(err, os.ErrNotExist) {
			log.Debug("sensor: monitor - no user info file")
		} else {
			log.Debugf("sensor: monitor - could not get user info file => %s", err)
		}
	}

	if dir := p.toDstPath(tmpDir); fsutil.DirExists(dir) {
		if err := os.Chmod(dir, os.ModeSticky|os.ModeDir|0777); err != nil {
			log.Warnf("saveArtifacts - error setting %s directory permission => %s", dir, err)
		}
	} else if err := os.MkdirAll(dir, 0777); err != nil {
		log.Warnf("saveArtifacts - error creating %s directory => %s", runDir, err)
	}

	if dir := p.toDstPath(runDir); !fsutil.DirExists(dir) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Warnf("saveArtifacts - error creating %s directory => %s", runDir, err)
		}
	}

	for srcDir := range extraDirs {
		dstDir := p.toDstPath(srcDir)
		if fsutil.DirExists(srcDir) && !fsutil.DirExists(dstDir) {
			if err := fsutil.CopyDirOnly(p.cmd.KeepPerms, srcDir, dstDir); err != nil {
				log.Warnf("CopyDirOnly(%v,%v) error => %v", srcDir, dstDir, err)
			}
		}
	}
}

func (p *artifactStore) toDstPath(path string) string {
	return fmt.Sprintf("%s/files%s", p.storeLocation, path)
}

func (p *artifactStore) copyDst(srcPath string, srcProps *report.ArtifactProps, newPerms map[string]*fsutil.AccessInfo) {
	dstPath := p.toDstPath(srcPath)

	if srcProps.Mode.IsDir() {
		perm := os.ModeSticky | os.ModeDir | 0777

		if !fsutil.DirExists(dstPath) {
			log.Debugf("saveArtifacts - creating dir %s -> %s", srcPath, dstPath)
			if err := os.MkdirAll(dstPath, perm); err != nil {
				log.Warnf("saveArtifacts - error creating %s directory => %s", dstPath, err)
			}
		} else {
			log.Debugf("saveArtifacts - changing dir perms (%s)", dstPath)
			if err := os.Chmod(dstPath, perm); err != nil {
				log.Warnf("saveArtifacts - error setting %s directory permission => %s", dstPath, err)
			}
		}

	} else {

		if fsutil.Exists(dstPath) {
			//we might already have the target file
			//when we have intermediate symlinks in the srcPath
			log.Debugf("saveArtifacts - target file already exists %s", dstPath)
		} else {

			var err error
			if isShellOrCmd(srcPath) {
				err = p.copyShell(srcPath)
			} else if isBin, _ := binfile.Detected(srcPath); isBin {
				err = p.copyBin(srcPath)
			} else {
				err = fsutil.CopyRegularFile(p.cmd.KeepPerms, srcPath, dstPath, true)
			}
			if err != nil {
				log.Warnf("saveArtifacts - error copying regular file => %s", err)
				return
			}
		}
	}

	if newPerms, hasNewPerms := newPerms[srcPath]; hasNewPerms {
		log.Debugf("saveArtifacts - setting new perms on copied file %s", dstPath)
		if err := fsutil.SetAccess(dstPath, newPerms); err != nil {
			log.Warnf("SetPerms(%v,%v) error => %v", dstPath, newPerms, err)
		}
	}
}

//TODO: review
func (p *artifactStore) trySymlink(linkName string, linkProps *report.ArtifactProps) bool {

	linkPath := p.toDstPath(linkName)
	linkDir := fsutil.FileDir(linkPath)
	//NOTE:
	//The symlink target dir might not exist, which means
	//the dir create calls that start with the current symlink prefix will fail.
	//We'll save the failed links to try again
	//later when the symlink target is already created.
	//Another option is to create the symlink targets,
	//but it might be tricky if the target is a symlink (potentially to another symlink, etc)

	if err := os.MkdirAll(linkDir, 0777); err != nil {
		log.Warnf("saveArtifacts.symlinkWalk - dir error (linkName=%s linkDir=%s linkPath=%s) => error=%v", linkName, linkDir, linkPath, err)
		return true
	}

	if linkProps.FSActivity != nil && linkProps.FSActivity.OpsCheckFile > 0 {
		log.Debugf("saveArtifacts.symlinkWalk - saving 'checked' symlink %s", linkName)
	}

	log.Debugf("saveArtifacts.symlinkWalk - symlinking file %s -> %s", linkProps.LinkRef, linkPath)
	if err := os.Symlink(linkProps.LinkRef, linkPath); err != nil {
		if errors.Is(err, os.ErrExist) {
			log.Debug("saveArtifacts.symlinkWalk - symlink already exists")
		} else {
			log.Warn("saveArtifacts.symlinkWalk - symlink create error => ", err)
		}
	}

	return false
}

func continueOn(path string, props *report.ArtifactProps, matcher *slimkeep.Matcher) bool {
	if matcher.MatchExclude(path, props.Mode.IsDir()) {
		log.Debugf("saveArtifacts - explicitly excluding %s", path)
		return true
	}

	if props.FSActivity == nil || props.FSActivity.OpsCheckFile == 0 {
		log.Debugf("saveArtifacts - saving file %s", path)
	} else {
		log.Debugf("saveArtifacts - saving 'checked' file %s", path)
		//NOTE: later have an option to save 'checked' only files without data
	}
	return false
}

func isShellOrCmd(path string) bool {
	base := filepath.Base(path)
	if base == "" {
		base = path
	}

	_, isShell := shellNames[base]
	_, isCmd := shellCommands[base]
	return isShell || isCmd
}

func (p *artifactStore) copyShell(srcPath string) error {
	shellPath, err := exec.LookPath(srcPath)
	if err != nil {
		log.Debugf("saveArtifacts - shell '%s' not found (%v)", srcPath, err)
		return err
	}

	return p.copyBin(shellPath)
}

func (p *artifactStore) copyBin(binPath string) error {
	binArtifacts, err := sodeps.AllDependencies(binPath)
	if err != nil {
		log.Warnf("saveArtifacts - %v - error getting bin artifacts => %v\n", binPath, err)
		binArtifacts = []string{binPath}
	}

	log.Debugf("saveArtifacts - include bin (%s, deps=%d): %s", binPath, len(binArtifacts), strings.Join(binArtifacts, ","))

	var lastErr error
	for _, bpath := range binArtifacts {
		dstPath := p.toDstPath(bpath)
		if fsutil.Exists(dstPath) {
			continue
		}
		if err := fsutil.CopyFile(p.cmd.KeepPerms, bpath, dstPath, true); err != nil {
			log.Warnf("CopyFile(%v,%v) error: %v", bpath, dstPath, err)
			lastErr = err
		}
	}
	return lastErr
}

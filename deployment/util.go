package deployment

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// Utility helper to convert relative paths to absolute.
func absPath(candidate string) string {
	s := candidate
	if !path.IsAbs(s) {
		if cwd, err := os.Getwd(); err == nil {
			s = path.Join(cwd, s)
		}
	}
	return s
}

// Utility helper to execute a system command.
func sysCommand(curDir string, command string, args []string) (string, string, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Dir = curDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Run()
	text := command + " " + strings.Join(args, " ")
	return text, curDir, strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String())
}

// Utility helper to create a directory and set its owner.
func makeDir(name string, uid, gid int, perm os.FileMode) error {
	if err := os.Mkdir(name, perm); err == nil {
		return setOwner(name, uid, gid)
	} else {
		return fmt.Errorf("unable to create directory: %s", err.Error())
	}
}

// Utility helper to set the owner of a file.
func setOwner(name string, uid, gid int) error {

	// We must be root, and we don't change ownership to root.
	if os.Geteuid() == 0 && uid != 0 && gid != 0 {
		if err := os.Chown(name, uid, gid); err != nil {
			return fmt.Errorf("unable to set owner: %s", err.Error())
		}
	}

	return nil
}

// Utility helper to set the owner of a directory subtree.
func setOwnerAll(dir string, uid, gid int) error {

	// We must be root, and we don't change ownership to root.
	if os.Geteuid() == 0 && uid != 0 && gid != 0 {

		// Visitor to do the actual work.
		var setOwnerFunc = func(path string, info os.FileInfo, err error) error {
			if err == nil {
				if err := os.Chown(path, uid, gid); err != nil {
					//fmt.Printf("unable to change owner to %d:%d: %s", uid, gid, path)
				}
			}
			return nil
		}

		filepath.Walk(dir, setOwnerFunc)
	}

	return nil
}

// Utility helper to generate an artifact filename and path.
func makeArtifactPath(dir, name, version, suffix string) (string, bool) {
	filename := fmt.Sprintf("%s-%s.%s", name, version, suffix)
	filepath := path.Join(dir, filename)
	if _, err := os.Stat(filepath); err == nil {
		return filepath, true
	} else {
		return filepath, false
	}
}

// Utility helper to generate an HMAC filename and path.
func makeHMACPath(dir, name, version, suffix string) (string, bool) {
	filename := fmt.Sprintf("%s-%s.%s.%s", name, version, suffix, kHMACSUFFIX)
	filepath := path.Join(dir, filename)
	if _, err := os.Stat(filepath); err == nil {
		return filepath, true
	} else {
		return filepath, false
	}
}

// Utility helper to generate a release dirname and path.
func makeReleasePath(dir, version string) (string, bool) {
	filepath := path.Join(dir, version)
	if _, err := os.Stat(filepath); err == nil {
		return filepath, true
	} else {
		return filepath, false
	}
}

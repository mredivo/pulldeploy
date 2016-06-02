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

// The values that are to be substituted into command line arguments.
type varValues struct {
	artifactPath string // var: #ARTIFACTPATH#
	versionDir   string // var: #VERSIONDIR#
}

// Utility helper to perform substitutions for supported command arguments.
func substituteVars(argsIn []string, values varValues) []string {
	var argsOut = make([]string, 0)
	for _, s := range argsIn {
		switch s {
		case "#ARTIFACTPATH#":
			argsOut = append(argsOut, values.artifactPath)
		case "#VERSIONDIR#":
			argsOut = append(argsOut, values.versionDir)
		default:
			argsOut = append(argsOut, s)
		}
	}
	return argsOut
}

// Utility helper to execute a system command.
func sysCommand(curDir string, command string, args []string) (string, error) {

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(command, args...)
	if curDir != "" {
		cmd.Dir = curDir
	} else {
		curDir, _ = os.Getwd()
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Format the results for easy logging.
	var logLine string
	var logErr error
	cmdline := command + " " + strings.Join(args, " ")
	if stdout.Len() == 0 {
		if stderr.Len() == 0 {
			logLine = fmt.Sprintf("Executed %q in %s", cmdline, curDir)
		} else {
			logLine = fmt.Sprintf("Executed %q in %s\nstderr=%q", cmdline, curDir,
				strings.TrimSpace(stderr.String()))
			if err != nil {
				logErr = fmt.Errorf("%s: %s", err.Error(), strings.TrimSpace(stderr.String()))
			} else {
				logErr = fmt.Errorf(strings.TrimSpace(stderr.String()))
			}
		}
	} else {
		if stderr.Len() == 0 {
			logLine = fmt.Sprintf("Executed %q in %s\nstdout=%q", cmdline, curDir,
				strings.TrimSpace(stdout.String()))
		} else {
			logLine = fmt.Sprintf("Executed %q in %s\nstdout=%q\nstderr=%q", cmdline, curDir,
				strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()))
			if err != nil {
				logErr = fmt.Errorf("%s: %s", err.Error(), strings.TrimSpace(stderr.String()))
			} else {
				logErr = fmt.Errorf(strings.TrimSpace(stderr.String()))
			}
		}
	}

	return logLine, logErr
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

	// We only do this as root, and we don't change ownership to root.
	if os.Geteuid() == 0 {
		if uid != 0 && gid != 0 {
			if err := os.Chown(name, uid, gid); err != nil {
				return fmt.Errorf("unable to set owner: %s", err.Error())
			}
		} else {
			return fmt.Errorf("refusing to set owner to %d:%d: %s", uid, gid, name)
		}
	}

	return nil
}

// Utility helper to set the owner of a directory subtree.
func setOwnerAll(dir string, uid, gid int) error {

	// We only do this as root, and we don't change ownership to root.
	if os.Geteuid() == 0 {

		errorCount := 0
		if uid != 0 && gid != 0 {

			// Visitor to do the actual work.
			var setOwnerFunc = func(path string, info os.FileInfo, err error) error {
				if err == nil {
					if err := os.Chown(path, uid, gid); err != nil {
						errorCount++
					}
				}
				return nil
			}

			// Visit every file and directory in the subtree.
			filepath.Walk(dir, setOwnerFunc)

			if errorCount > 0 {
				return fmt.Errorf("unable to change owner to %d:%d: for %d file(s)", uid, gid, errorCount)
			}
		} else {
			return fmt.Errorf("refusing to set owner to %d:%d: %s", uid, gid, dir)
		}
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

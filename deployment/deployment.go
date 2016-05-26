/*
Package deployment provides methods for managing application deployment and release files.

A deployment resides on the server running PullDeploy in daemon mode. It has the following
directory structure:

	/BASEDIR/APPNAME/artifact
	/BASEDIR/APPNAME/current  (a symlink)
	/BASEDIR/APPNAME/release

Artifacts retrieved from the repository are placed into the "artifact" directory:

	/BASEDIR/APPNAME/artifact/APPNAME-VERSION.ARTIFACTTYPE

Deployed releases are unpacked into a directory named for the version, under the
"release" directory.

	/BASEDIR/APPNAME/release/VERSION1
	/BASEDIR/APPNAME/release/VERSION2
	/BASEDIR/APPNAME/release/VERSION3

Releasing a version points the "current" symlink to the specified release directory.

*/
package deployment

import (
	"crypto/hmac"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"

	"github.com/mredivo/pulldeploy/pdconfig"
)

const kARTIFACTDIR = "artifact"
const kRELEASEDIR = "release"
const kCURRENTDIR = "current"
const kHMACSUFFIX = "hmac"

// Deployment provides methods for manipulating local deployment files.
type Deployment struct {
	appName     string                  // The name of the application
	cfg         pdconfig.AppConfig      // The deployment configuration
	acfg        pdconfig.ArtifactConfig // The Artifact Type configuration
	uid         int                     // The numeric UID to own all files for this deployment
	gid         int                     // The numeric GID to own all files for this deployment
	baseDir     string                  // The derived top-level directory for this app's files
	artifactDir string                  // The derived subdirectory for fetched build artifacts
	releaseDir  string                  // The derived subdirectory for extracted build artifacts
}

// New returns a new Deployment.
func New(appName string, pdcfg pdconfig.PDConfig, cfg *pdconfig.AppConfig) (*Deployment, error) {

	d := new(Deployment)
	d.cfg = *cfg

	// Capture the supplied arguments.
	d.appName = appName

	// All string arguments are mandatory.
	if appName == "" {
		return nil, errors.New("Deployment initialization error: Appname is mandatory")
	}
	if d.cfg.BaseDir == "" {
		return nil, errors.New("Deployment initialization error: BaseDir is mandatory")
	}

	// Validate the artifact type.
	if ac, err := pdcfg.GetArtifactConfig(d.cfg.ArtifactType); err == nil {
		d.acfg = *ac
	} else {
		return nil, fmt.Errorf("Deployment initialization error: invalid ArtifactType %q", d.cfg.ArtifactType)
	}

	// Derive the UID/GID from the username/groupname.
	// NOTE: Go doesn't yet support looking up a GID from a name, so
	//       we use the gid from the user.
	if user, err := user.Lookup(d.cfg.User); err == nil {
		if i, err := strconv.ParseInt(user.Uid, 10, 64); err == nil {
			d.uid = int(i)
		}
		if i, err := strconv.ParseInt(user.Gid, 10, 64); err == nil {
			d.gid = int(i)
		}
	}

	// The parent directory must not be "/".
	parentDir := absPath(d.cfg.BaseDir)
	if parentDir == "/" {
		return nil, errors.New("Deployment initialization error: \"/\" not permitted as BaseDir")
	}

	// The parent directory must exist.
	if _, err := os.Stat(parentDir); err != nil {
		return nil, fmt.Errorf("Deployment initialization error: unable to stat BaseDir: %s", err.Error())
	}

	// If the base dir doesn't exist, create it.
	d.baseDir = path.Join(parentDir, appName)
	if _, err := os.Stat(d.baseDir); err != nil {
		if err := makeDir(d.baseDir, d.uid, d.gid, 0755); err != nil {
			return nil, fmt.Errorf("Deployment initialization error: %s", err.Error())
		}
	}

	// If the artifact dir doesn't exist, create it.
	d.artifactDir = path.Join(d.baseDir, kARTIFACTDIR)
	if _, err := os.Stat(d.artifactDir); err != nil {
		if err := makeDir(d.artifactDir, d.uid, d.gid, 0755); err != nil {
			return nil, fmt.Errorf("Deployment initialization error: %s", err.Error())
		}
	}

	// If the release dir doesn't exist, create it.
	d.releaseDir = path.Join(d.baseDir, kRELEASEDIR)
	if _, err := os.Stat(d.releaseDir); err != nil {
		if err := makeDir(d.releaseDir, d.uid, d.gid, 0755); err != nil {
			return nil, fmt.Errorf("Deployment initialization error: %s", err.Error())
		}
	}

	return d, nil
}

// ArtifactPresent indicates whether the artifact has already been written.
func (d *Deployment) ArtifactPresent(version string) bool {

	// Generate the filename, and check whether file already exists.
	_, exists := makeArtifactPath(d.artifactDir, d.appName, version, d.acfg.Extension)
	return exists
}

// Write creates a file in the artifact area from the given stream.
func (d *Deployment) WriteArtifact(version string, rc io.ReadCloser) error {

	// Housekeeping: ensure the source is closed when done.
	defer rc.Close()

	// Generate the filename, and check whether file already exists.
	artifactPath, exists := makeArtifactPath(d.artifactDir, d.appName, version, d.acfg.Extension)
	if exists {
		return fmt.Errorf("Artifact already exists: %s", artifactPath)
	}

	// Open the file, and write the data into it.
	if fp, err := os.OpenFile(artifactPath, os.O_WRONLY|os.O_CREATE, 0664); err == nil {
		defer fp.Close()
		if _, err := io.Copy(fp, rc); err != nil {
			return fmt.Errorf("Error while creating %q: %s", artifactPath, err.Error())
		}
		if err := setOwner(artifactPath, d.uid, d.gid); err != nil {
			return fmt.Errorf("Unable to set owner on %q: %s", artifactPath, err.Error())
		}
	}

	return nil
}

// HMACPresent indicates whether the HMAC has already been written.
func (d *Deployment) HMACPresent(version string) bool {

	// Generate the filename, and check whether file already exists.
	_, exists := makeHMACPath(d.artifactDir, d.appName, version, d.acfg.Extension)
	return exists
}

// WriteHMAC writes an HMAC into the artifact area.
func (d *Deployment) WriteHMAC(version string, hmac []byte) error {

	// Generate the filename, write to file, set ownership.
	hmacPath, _ := makeHMACPath(d.artifactDir, d.appName, version, d.acfg.Extension)
	if err := ioutil.WriteFile(hmacPath, hmac, 0664); err != nil {
		return fmt.Errorf("Error while writing %q: %s", hmacPath, err.Error())
	}
	if err := setOwner(hmacPath, d.uid, d.gid); err != nil {
		return fmt.Errorf("Unable to set owner on %q: %s", hmacPath, err.Error())
	}

	return nil
}

// CheckHMAC confirms that the artifact has not been corrupted or tampered with by
// calculating its HMAC and comparing it with the retrieved HMAC.
func (d *Deployment) CheckHMAC(version string) error {

	// Build the filenames.
	artifactPath, exists := makeArtifactPath(d.artifactDir, d.appName, version, d.acfg.Extension)
	if !exists {
		return fmt.Errorf("Artifact does not exist: %s", artifactPath)
	}
	hmacPath, exists := makeHMACPath(d.artifactDir, d.appName, version, d.acfg.Extension)
	if !exists {
		return fmt.Errorf("HMAC does not exist: %s", artifactPath)
	}

	// Read in the HMAC.
	if expectedMAC, err := ioutil.ReadFile(hmacPath); err == nil {

		// Open the artifact, and calculate its HMAC.
		if fp, err := os.Open(artifactPath); err == nil {
			messageMAC := CalculateHMAC(fp, NewHMACCalculator(d.cfg.Secret))
			if !hmac.Equal(messageMAC, expectedMAC) {
				return fmt.Errorf(
					"Artifact is corrupt: Expected HMAC: %q: Calculated HMAC: %q",
					string(expectedMAC),
					string(messageMAC),
				)
			}

		} else {
			return fmt.Errorf("Error while reading %q: %s", artifactPath, err.Error())
		}
	} else {
		return fmt.Errorf("Error while reading %q: %s", hmacPath, err.Error())
	}

	return nil
}

// Extract transfers an artifact to the version release directory.
func (d *Deployment) Extract(version string) error {

	// Ensure that the artifact to be extracted exists.
	artifactPath, exists := makeArtifactPath(d.artifactDir, d.appName, version, d.acfg.Extension)
	if !exists {
		return fmt.Errorf("Artifact does not exist: %s", artifactPath)
	}

	// Ensure the extract command wasn't loaded from an insecure file.
	if os.Geteuid() == 0 && d.acfg.Insecure {
		return fmt.Errorf(
			"Refusing to execute extract command from insecure \"pulldeploy.yaml\" as root")
	}

	// Create the version directory if it doesn't exist.
	versionDir, exists := makeReleasePath(d.releaseDir, version)
	if !exists {
		if err := makeDir(versionDir, d.uid, d.gid, 0755); err != nil {
			return fmt.Errorf("Cannot create release directory %q: %s", version, err.Error())
		}
	}

	// Build the argument list for the extract command.
	var extractArgs = make([]string, 0)
	for _, s := range d.acfg.Extract.Args {
		switch s {
		case "#ARTIFACTPATH#":
			extractArgs = append(extractArgs, artifactPath)
		case "#VERSIONDIR#":
			extractArgs = append(extractArgs, versionDir)
		default:
			extractArgs = append(extractArgs, s)
		}
	}

	// Extract the archive into the version directory.
	_, err := sysCommand("", d.acfg.Extract.Cmd, extractArgs)
	if err != nil {
		return fmt.Errorf("Cannot extract archive %q into %q: %s", artifactPath, versionDir, err.Error())
	}

	// Set the ownership of all the extracted files.
	if err := setOwnerAll(versionDir, d.uid, d.gid); err != nil {
		return err
	}

	return nil
}

// Link sets the "current" symlink to point at the indicated version.
func (d *Deployment) Link(version string) error {
	versionDir, exists := makeReleasePath(d.releaseDir, version)
	if !exists {
		return fmt.Errorf("Release directory does not exist: %q", versionDir)
	}
	symlinkPath := path.Join(d.baseDir, kCURRENTDIR)
	os.Remove(symlinkPath)
	return os.Symlink(versionDir, symlinkPath)
}

// PostDeploy executes the configured PostDeploy command.
func (d *Deployment) PostDeploy(version string) (string, error) {
	if os.Geteuid() == 0 && d.cfg.Insecure {
		return "", fmt.Errorf(
			"Refusing to execute post-deploy command from insecure %q configuration as root",
			d.appName)
	}
	if d.cfg.Scripts["postdeploy"].Cmd != "" {
		versionDir, _ := makeReleasePath(d.releaseDir, version)
		return sysCommand(versionDir, d.cfg.Scripts["postdeploy"].Cmd, d.cfg.Scripts["postdeploy"].Args)
	}
	return "", nil
}

// PostRelease executes the configured PostRelease command.
func (d *Deployment) PostRelease(version string) (string, error) {
	if os.Geteuid() == 0 && d.cfg.Insecure {
		return "", fmt.Errorf(
			"Refusing to execute post-release command from insecure %q configuration as root",
			d.appName)
	}
	if d.cfg.Scripts["postrelease"].Cmd != "" {
		versionDir, _ := makeReleasePath(d.releaseDir, version)
		return sysCommand(versionDir, d.cfg.Scripts["postrelease"].Cmd, d.cfg.Scripts["postrelease"].Args)
	}
	return "", nil
}

// Remove deletes everything associated with the given name.
func (d *Deployment) Remove(version string) error {

	// Removing the currently linked version is not permitted.
	if d.GetCurrentLink() == version {
		return fmt.Errorf("Removing current version not permitted: %q", version)
	}

	// Remove the artifact and HMAC.
	if artifactPath, exists := makeArtifactPath(d.artifactDir, d.appName, version, d.acfg.Extension); exists {
		os.Remove(artifactPath)
	}
	if hmacPath, exists := makeHMACPath(d.artifactDir, d.appName, version, d.acfg.Extension); exists {
		os.Remove(hmacPath)
	}

	// Remove the extracted files.
	if versionDir, exists := makeReleasePath(d.releaseDir, version); exists {
		return os.RemoveAll(versionDir)
	}

	return nil
}

// GetCurrentLink returns the name of the currently released version.
func (d *Deployment) GetCurrentLink() string {

	// Read the symlink off disk.
	symlink, err := os.Readlink(path.Join(d.baseDir, kCURRENTDIR))
	if err != nil {
		return ""
	}

	// We are interested in the last element, which is the active version directory.
	dirs := strings.Split(symlink, "/")
	return dirs[len(dirs)-1]
}

// ListVersions enumerates all the versions currently available for linking.
func (d *Deployment) ListVersions() []string {

	var versionList []string

	// Everything in the release directory is an available version.
	if fi, err := ioutil.ReadDir(d.releaseDir); err == nil {
		for _, v := range fi {
			versionList = append(versionList, v.Name())
		}
	}

	return versionList
}

package storage

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

/*
Repository data is stored in the local filesystem.

Valid Params for KST_LOCAL:

	* "basedir" The full path to the directory containing the repository
*/
const KST_LOCAL StorageType = "local"

// stLocal is used for PullDeploy repositories on the local filesystem.
type stLocal struct {
	baseDir string // The root directory of the repo in the local filesystem
}

// Initialize the repository object.
func (st *stLocal) init(params Params) error {
	if baseDir, ok := params["basedir"]; ok {
		st.baseDir = absPath(baseDir)
		if _, err := os.Stat(st.baseDir); err != nil {
			return fmt.Errorf("Storage initialization error: basedir: %s", err.Error())
		}
	} else {
		return fmt.Errorf("Storage initialization error: %q is a required parameter", "basedir")
	}
	return nil
}

// Get fetches the contents of a repository file into a byte array.
func (st *stLocal) Get(repoPath string) ([]byte, error) {

	// Generate the filename, and check that path exists.
	fullPath, exists := makeLocalPath(st.baseDir, repoPath)
	if !exists {
		return []byte{}, fmt.Errorf("Not found: %s", fullPath)
	}

	return ioutil.ReadFile(fullPath)
}

// Put writes the contents of a byte array into a repository file.
func (st *stLocal) Put(repoPath string, data []byte) error {

	// Generate the filename, and ensure path exists.
	fullPath, exists := makeLocalPath(st.baseDir, repoPath)
	if !exists {
		if err := os.MkdirAll(path.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("Error while creating %q: %s", fullPath, err.Error())
		}
	}

	return ioutil.WriteFile(fullPath, data, 0644)
}

// GetReader returns a stream handle for reading a repository file.
func (st *stLocal) GetReader(repoPath string) (io.ReadCloser, error) {

	// Generate the filename, and check that path exists.
	fullPath, exists := makeLocalPath(st.baseDir, repoPath)
	if !exists {
		return nil, fmt.Errorf("Not found: %s", fullPath)
	}

	return os.Open(fullPath)
}

// PutReader writes a stream to a repository file.
func (st *stLocal) PutReader(repoPath string, rc io.ReadCloser, length int64) error {

	// Housekeeping: ensure the source is closed when done.
	defer rc.Close()

	// Generate the filename, and ensure path exists.
	fullPath, exists := makeLocalPath(st.baseDir, repoPath)
	if !exists {
		if err := os.MkdirAll(path.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("Error while creating %q: %s", fullPath, err.Error())
		}
	}

	// Open the file, and write the data into it.
	if fp, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, 0664); err == nil {
		defer fp.Close()
		if _, err := io.Copy(fp, rc); err != nil {
			return fmt.Errorf("Error while creating %q: %s", fullPath, err.Error())
		}
	}

	return nil
}

// Utility helper to generate a local repository full path.
func makeLocalPath(baseDir, repoPath string) (string, bool) {
	fullpath := path.Join(baseDir, repoPath)
	if _, err := os.Stat(fullpath); err == nil {
		return fullpath, true
	} else {
		return fullpath, false
	}
}

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

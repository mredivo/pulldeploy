// Package repo defines the objects stored in the repository.
package repo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

// Version provides version and filename information for uploaded files.
type Version struct {
	Name     string `json:"version"`  // The name of the version, expected to be similar to "1.0.0"
	Filename string `json:"filename"` // The name of the uploaded tar file for this version
	Released bool   `json:"released"` // True if this version has ever been released
	Enabled  bool   `json:"enabled"`  // False if this version has been specifically disabled; default True
}

// Env enumerates the versions deployed to an environment, and identifies the current release.
type Env struct {
	Prev       string   `json:"prev"`       // The index into the Deployed map of a candidate version
	Current    string   `json:"current"`    // The index into the Deployed map of the current version
	Next       string   `json:"next"`       // The index into the Deployed map of a candidate version
	Deployed   []string `json:"deployed"`   // The set of versions deployed to this environment
	Released   []string `json:"released"`   // The set of versions released to this environment
	Previewers []string `json:"previewers"` // An array of hostnames eligible for the Next version
}

// RepoIndex is the repository index for an application.
type RepoIndex struct {
	appName  string             // The name of the application in this index
	Canary   int                `json:"canary"`       // Incremented each time the index is written out
	Keep     int                `json:"keep"`         // The minimum number of versions to retain when purging
	Versions map[string]Version `json:"versions"`     // The set of versions uploaded; old entries fall off
	Envs     map[string]Env     `json:"environments"` // The defined environments: prod, stage, etc.
}

// NewRepoIndex returns a new instance of RepoIndex.
func NewRepoIndex(appName string) *RepoIndex {

	ri := new(RepoIndex)

	ri.appName = strings.ToLower(appName)
	ri.Canary = 0
	ri.Keep = 5
	ri.Versions = make(map[string]Version)
	ri.Envs = make(map[string]Env)

	return ri
}

// AddEnv inserts a new environment into the index.
func (ri *RepoIndex) AddEnv(envName string) error {
	if _, found := ri.Envs[envName]; found {
		return fmt.Errorf("environment %q already present", envName)
	}
	ri.Envs[envName] = Env{Deployed: []string{}, Released: []string{}, Previewers: []string{}}
	return nil
}

// AddVersion inserts a new version into the index.
func (ri *RepoIndex) AddVersion(name, filename string, enabled bool) error {
	if _, found := ri.Versions[name]; found {
		return fmt.Errorf("version %q already present", name)
	}
	ri.Versions[name] = Version{Name: name, Filename: filename, Released: false, Enabled: enabled}
	return nil
}

// RepoIndexPath returns the canonical path to the app's index in the repository.
func (ri *RepoIndex) RepoIndexPath() string {
	return ri.appName + "/index.json"
}

// RepoArtifactPath returns the canonical path to the indicated artifact.
func (ri *RepoIndex) RepoArtifactPath(filename string) string {
	return path.Join(ri.appName, "versions", filename)
}

// RepoArtifactFilename returns the canonical filename of the indicated artifact.
func (ri *RepoIndex) RepoArtifactFilename(version, filename string) string {
	parts := strings.Split(filename, ".")
	ext := strings.Join(parts[1:], ".")
	return ri.appName + "-" + version + "." + ext
}

// FromJSON materializes the index from a JSON byte array.
func (ri *RepoIndex) FromJSON(text []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(text))
	return decoder.Decode(ri)
}

// ToJSON serializes the index to a JSON byte array.
func (ri *RepoIndex) ToJSON() ([]byte, error) {
	if text, err := json.MarshalIndent(*ri, "", "    "); err == nil {
		return text, nil
	} else {
		return []byte{}, err
	}
}

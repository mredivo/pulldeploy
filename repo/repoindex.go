// Package repo defines the objects stored in the repository.
package repo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

// RepoIndex is the repository index for an application.
type RepoIndex struct {
	appName  string              // The name of the application in this index
	Canary   int                 `json:"canary"`       // Incremented each time the index is written out
	Versions map[string]*Version `json:"versions"`     // The set of versions uploaded; old entries fall off
	Envs     map[string]*Env     `json:"environments"` // The defined environments: prod, stage, etc.
}

// NewRepoIndex returns a new instance of RepoIndex.
func NewRepoIndex(appName string) *RepoIndex {

	ri := new(RepoIndex)

	ri.appName = strings.ToLower(appName)
	ri.Canary = 0
	ri.Versions = make(map[string]*Version)
	ri.Envs = make(map[string]*Env)

	return ri
}

// AddEnv initializes a new environment and adds it into the index.
func (ri *RepoIndex) AddEnv(envName string) error {
	if _, err := ri.GetEnv(envName); err == nil {
		return fmt.Errorf("environment %q already present", envName)
	}
	ri.SetEnv(envName, newEnv())
	return nil
}

// GetEnv retrieves an environment from the index.
func (ri *RepoIndex) GetEnv(envName string) (*Env, error) {
	if env, found := ri.Envs[envName]; found {
		return env, nil
	} else {
		return nil, fmt.Errorf("environment %q not present", envName)
	}
}

// SetEnv replaces an environment in the index.
func (ri *RepoIndex) SetEnv(envName string, env *Env) error {
	ri.Envs[envName] = env
	return nil
}

// RmEnv removes an environment from the index.
func (ri *RepoIndex) RmEnv(envName string) error {
	if _, err := ri.GetEnv(envName); err != nil {
		return err
	}
	delete(ri.Envs, envName)
	return nil
}

// AddVersion initializes a new version and adds it into the index.
func (ri *RepoIndex) AddVersion(versionName, fileName string, enabled bool) error {
	if _, err := ri.GetVersion(versionName); err == nil {
		return fmt.Errorf("version %q already present", versionName)
	}
	ri.SetVersion(versionName, newVersion(versionName, fileName, enabled))
	return nil
}

// GetVersion retrieves a version from the index.
func (ri *RepoIndex) GetVersion(versionName string) (*Version, error) {
	if version, found := ri.Versions[versionName]; found {
		return version, nil
	} else {
		return nil, fmt.Errorf("version %q not present", versionName)
	}
}

// SetVersion replaces a version in the index.
func (ri *RepoIndex) SetVersion(versionName string, version *Version) error {
	ri.Versions[versionName] = version
	return nil
}

// RmVersion removes a version from the index.
func (ri *RepoIndex) RmVersion(versionName string) error {
	if _, err := ri.GetVersion(versionName); err != nil {
		return err
	}
	delete(ri.Versions, versionName)
	return nil
}

// IndexPath returns the canonical path to the app's index in the repository.
func (ri *RepoIndex) IndexPath() string {
	return ri.appName + "/index.json"
}

// ArtifactPath returns the canonical path to the indicated artifact.
func (ri *RepoIndex) ArtifactPath(filename string) string {
	return path.Join(ri.appName, "versions", filename)
}

// SignatureFilename returns the signature filename for the indicated artifact.
func (ri *RepoIndex) SignaturePath(filename string) string {
	return ri.ArtifactPath(filename) + ".hmac"
}

// ArtifactFilename returns the canonical filename of the indicated artifact.
func (ri *RepoIndex) ArtifactFilename(version, filename string) string {
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

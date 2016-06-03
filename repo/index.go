// Package repo defines the Pulldeploy repository metadata.
package repo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

// The type of the callback called when versions age out of the repo.
type versionOnDelete func(versionName string)

// Index is the repository index for an application.
type Index struct {
	appName  string              // The name of the application in this index
	Canary   int                 `json:"canary"`       // Incremented each time the index is written out
	Versions map[string]*Version `json:"versions"`     // The set of versions uploaded; old entries fall off
	Envs     map[string]*Env     `json:"environments"` // The defined environments: prod, stage, etc.
}

// NewIndex returns a new instance of Index.
func NewIndex(appName string) *Index {

	ri := new(Index)

	ri.appName = strings.ToLower(appName)
	ri.Canary = 0
	ri.Versions = make(map[string]*Version)
	ri.Envs = make(map[string]*Env)

	return ri
}

// AddEnv initializes a new environment and adds it into the index.
func (ri *Index) AddEnv(envName string) error {
	if _, err := ri.GetEnv(envName); err == nil {
		return fmt.Errorf("environment %q already present", envName)
	}
	ri.SetEnv(envName, newEnv())
	return nil
}

// GetEnv retrieves an environment from the index.
func (ri *Index) GetEnv(envName string) (*Env, error) {
	if env, found := ri.Envs[envName]; found {
		// Provide private access to the Versions.
		env.versions = ri.Versions
		return env, nil
	}
	return nil, fmt.Errorf("environment %q not present", envName)
}

// SetEnv replaces an environment in the index.
func (ri *Index) SetEnv(envName string, env *Env) error {
	ri.Envs[envName] = env
	return nil
}

// RmEnv removes an environment from the index.
func (ri *Index) RmEnv(envName string) error {
	if _, err := ri.GetEnv(envName); err != nil {
		return err
	}
	delete(ri.Envs, envName)
	return nil
}

// AddVersion initializes a new version and adds it into the index.
func (ri *Index) AddVersion(versionName, fileName string, enabled bool, onDelete versionOnDelete) error {

	if _, err := ri.GetVersion(versionName); err == nil {
		return fmt.Errorf("version %q already present", versionName)
	}

	// Determine the minimum number of version entries that must be kept.
	minCount := 0
	for _, env := range ri.Envs {
		if env.Keep > minCount {
			minCount = env.Keep
		}
	}

	// Get versions, oldest first, and determine how many we currently have.
	versions := ri.VersionList("asc")
	curCount := len(versions)

	// Remove unreferenced versions until we reach the minimum count.
	for _, vers := range versions {
		if curCount < minCount {
			break
		}

		// Check for references in each environment.
		referenced := false
		for _, env := range ri.Envs {
			for _, histEvent := range env.Deployed {
				if vers.Name == histEvent.Version {
					referenced = true
					break
				}
			}
			if referenced {
				break
			}
		}

		// If unreferenced, remove it from the list.
		if !referenced {
			onDelete(vers.Name)
			delete(ri.Versions, vers.Name)
			curCount--
		}
	}

	// Now add the new version (which is of course unreferenced so far).
	ri.SetVersion(versionName, newVersion(versionName, fileName, enabled))

	return nil
}

// GetVersion retrieves a version from the index.
func (ri *Index) GetVersion(versionName string) (*Version, error) {
	if version, found := ri.Versions[versionName]; found {
		return version, nil
	}
	return nil, fmt.Errorf("version %q not present", versionName)
}

// SetVersion replaces a version in the index.
func (ri *Index) SetVersion(versionName string, version *Version) error {
	ri.Versions[versionName] = version
	return nil
}

// RmVersion removes a version from the index and from all environments.
func (ri *Index) RmVersion(versionName string) error {

	// The version must be present.
	if _, err := ri.GetVersion(versionName); err != nil {
		return err
	}

	// Query every environment to determine whether the version is in use.
	usedBy := make([]string, 0)
	for envName, env := range ri.Envs {
		if !env.isPurgable(versionName) {
			usedBy = append(usedBy, envName)
		}
	}
	if len(usedBy) > 0 {
		return fmt.Errorf("Version %q in use in %s", versionName, strings.Join(usedBy, ", "))
	}

	// Remove the version from all environments.
	errCount := 0
	for _, env := range ri.Envs {
		if err := env.purgeVersion(versionName); err != nil {
			errCount++
		}
	}

	// Remove the version from the versions list.
	if errCount == 0 {
		delete(ri.Versions, versionName)
	} else {
		return fmt.Errorf("Version %q not completely purged", versionName)
	}
	return nil
}

// VersionList returns an array of versions ordered by timestamp.
func (ri *Index) VersionList(order string) []Version {
	var versions []Version
	for _, v := range ri.Versions {
		versions = append(versions, *v)
	}
	ascending := func(v1, v2 *Version) bool {
		return v2.TS.After(v1.TS)
	}
	descending := func(v1, v2 *Version) bool {
		return v1.TS.After(v2.TS)
	}
	if order == "desc" {
		sortVersionsBy(descending).Sort(versions)
	} else {
		sortVersionsBy(ascending).Sort(versions)
	}
	return versions
}

// IndexPath returns the canonical path to the app's index in the repository.
func (ri *Index) IndexPath() string {
	return ri.appName + "/index.json"
}

// ArtifactPath returns the canonical path to the indicated artifact.
func (ri *Index) ArtifactPath(filename string) string {
	return path.Join(ri.appName, "versions", filename)
}

// HMACPath returns the canonical path to the HMAC for the indicated artifact.
func (ri *Index) HMACPath(filename string) string {
	return ri.ArtifactPath(filename) + ".hmac"
}

// ArtifactFilename returns the canonical filename of the indicated artifact.
func (ri *Index) ArtifactFilename(version, artifactType string) string {
	return ri.appName + "-" + version + "." + artifactType
}

// FromJSON materializes the index from a JSON byte array.
func (ri *Index) FromJSON(text []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(text))
	return decoder.Decode(ri)
}

// ToJSON serializes the index to a JSON byte array.
func (ri *Index) ToJSON() ([]byte, error) {
	if text, err := json.MarshalIndent(*ri, "", "    "); err == nil {
		return text, nil
	} else {
		return []byte{}, err
	}
}

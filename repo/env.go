package repo

import (
	"fmt"
)

// Env enumerates the versions deployed to an environment, and identifies the current release.
type Env struct {
	Keep       int      `json:"keep"`       // The maximum number of versions to retain when adding
	Prev       string   `json:"prev"`       // The index into the Deployed map of a candidate version
	Current    string   `json:"current"`    // The index into the Deployed map of the current version
	Next       string   `json:"next"`       // The index into the Deployed map of a candidate version
	Deployed   []string `json:"deployed"`   // The set of versions deployed to this environment
	Released   []string `json:"released"`   // The set of versions released to this environment
	Previewers []string `json:"previewers"` // An array of hostnames eligible for the Next version
}

func newEnv() *Env {
	return &Env{Keep: 5, Deployed: []string{}, Released: []string{}, Previewers: []string{}}
}

func (env *Env) SetKeep(keep int) {
	env.Keep = keep
}

type deployOnDelete func(versionName string)

// Deploy makes an uploaded artifact available in this environment.
func (env *Env) Deploy(versionName string, onDelete deployOnDelete) error {

	// Ensure that this version of the artifact is not already deployed.
	for _, v := range env.Deployed {
		if v == versionName {
			return fmt.Errorf("version %q already deployed", versionName)
		}
	}

	// Ensure the number of entries will not exceed the cap.
	for entryCount := len(env.Deployed); entryCount >= env.Keep; entryCount-- {
		// Do not remove the current or immediately prior releases.
		candidate := entryCount - 1
		if candidate >= 0 {
			if env.Deployed[candidate] == env.Current || env.Deployed[candidate] == env.Prev {
				candidate-- // Skip this one
			}
		}
		if candidate >= 0 {
			if env.Deployed[candidate] == env.Current || env.Deployed[candidate] == env.Prev {
				candidate-- // Skip this one too
			}
		}
		if candidate >= 0 {
			onDelete(env.Deployed[candidate])
			env.Deployed = append(env.Deployed[:candidate], env.Deployed[candidate+1:]...)
		}
	}

	// Add the new entry at the beginning of the list.
	env.Deployed = append([]string{versionName}, env.Deployed...)

	return nil
}

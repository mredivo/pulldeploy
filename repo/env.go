package repo

import (
	"fmt"
)

const kMAX_RLS_HST_ENTRIES = 10

// Env enumerates the versions deployed to an environment, and identifies the current release.
type Env struct {
	Keep       int      `json:"keep"`       // The maximum number of versions to retain when adding
	Prior      string   `json:"prior"`      // The version most recently active prior to current
	Current    string   `json:"current"`    // The currently active version
	Preview    string   `json:"preview"`    // The version considered active by the Previewers hosts
	Deployed   []string `json:"deployed"`   // The set of versions deployed to this environment
	Released   []string `json:"released"`   // The set of versions released to this environment
	Previewers []string `json:"previewers"` // The set of hostnames eligible for the Preview version
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

	// If over the cap, attempt to reduce the number of entries.
	for entryCount := len(env.Deployed); entryCount >= env.Keep; entryCount-- {
		// Do not remove the current, prior, or preview releases.
		candidate := entryCount - 1
		if candidate >= 0 {
			if env.Deployed[candidate] == env.Current ||
				env.Deployed[candidate] == env.Prior ||
				env.Deployed[candidate] == env.Preview {
				candidate-- // Skip this one
			}
		}
		if candidate >= 0 {
			if env.Deployed[candidate] == env.Current ||
				env.Deployed[candidate] == env.Prior ||
				env.Deployed[candidate] == env.Preview {
				candidate-- // Skip this one too
			}
		}
		if candidate >= 0 {
			if env.Deployed[candidate] == env.Current ||
				env.Deployed[candidate] == env.Prior ||
				env.Deployed[candidate] == env.Preview {
				candidate-- // And this one
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

// Release makes a deployed artifact the currently active one in this environment.
func (env *Env) Release(versionName string, previewers []string) error {

	// Ensure that this version of the artifact has been deployed.
	found := false
	for _, v := range env.Deployed {
		if v == versionName {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("version %q not deployed", versionName)
	}

	// Determine whether this is a general release, or just a preview.
	if len(previewers) > 0 {
		return env.releasePreview(versionName, previewers)
	} else {
		return env.releaseGeneral(versionName)
	}
}

func (env *Env) releaseGeneral(versionName string) error {

	// A general release cancels any outstanding preview.
	env.Preview = ""
	env.Previewers = []string{}

	// Establish the current and previous versions.
	if env.Current != versionName {

		env.Prior = env.Current
		env.Current = versionName

		// Append to release history, and remove old entries when size maxes out.
		env.Released = append([]string{versionName}, env.Released...)
		if len(env.Released) > kMAX_RLS_HST_ENTRIES {
			env.Released = env.Released[:kMAX_RLS_HST_ENTRIES]
		}
	}

	return nil
}

func (env *Env) releasePreview(versionName string, previewers []string) error {
	// This has no effect on anything but the previewers.
	env.Preview = versionName
	env.Previewers = previewers
	return nil
}

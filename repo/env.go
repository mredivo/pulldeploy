package repo

import (
	"fmt"
	"time"
)

const kMAX_RLS_HST_ENTRIES = 10

// HistEvent associates a timestamp with a version for deploy/release activity.
type HistEvent struct {
	Version string    `json:"version"`   // The version affected
	TS      time.Time `json:"timestamp"` // The time at which the event occurred
}

// Env enumerates the versions deployed to an environment, and identifies the current release.
type Env struct {
	Keep       int         `json:"keep"`       // The maximum number of versions to retain when adding
	Prior      string      `json:"prior"`      // The version most recently active prior to current
	Current    string      `json:"current"`    // The currently active version
	Preview    string      `json:"preview"`    // The version considered active by the Previewers hosts
	Deployed   []HistEvent `json:"deployed"`   // The set of versions deployed to this environment
	Released   []HistEvent `json:"released"`   // The set of versions released to this environment
	Previewers []string    `json:"previewers"` // The set of hostnames eligible for the Preview version
	versions   map[string]*Version
}

func newEnv() *Env {
	return &Env{Keep: 5, Deployed: []HistEvent{}, Released: []HistEvent{}, Previewers: []string{}}
}

// SetKeep sets the number of versions to keep in the repo and on the servers.
func (env *Env) SetKeep(keep int) {
	env.Keep = keep
}

type deployOnDelete func(versionName string)

// Deploy makes an uploaded artifact available in this environment.
func (env *Env) Deploy(versionName string, onDelete deployOnDelete) error {

	// Ensure that this version of the artifact is not already deployed.
	for _, v := range env.Deployed {
		if v.Version == versionName {
			return fmt.Errorf("version %q already deployed", versionName)
		}
	}

	// If over the cap, attempt to reduce the number of entries.
	for entryCount := len(env.Deployed); entryCount >= env.Keep; entryCount-- {
		// Do not remove the current, prior, or preview releases.
		candidate := entryCount - 1
		if candidate >= 0 {
			if env.Deployed[candidate].Version == env.Current ||
				env.Deployed[candidate].Version == env.Prior ||
				env.Deployed[candidate].Version == env.Preview {
				candidate-- // Skip this one
			}
		}
		if candidate >= 0 {
			if env.Deployed[candidate].Version == env.Current ||
				env.Deployed[candidate].Version == env.Prior ||
				env.Deployed[candidate].Version == env.Preview {
				candidate-- // Skip this one too
			}
		}
		if candidate >= 0 {
			if env.Deployed[candidate].Version == env.Current ||
				env.Deployed[candidate].Version == env.Prior ||
				env.Deployed[candidate].Version == env.Preview {
				candidate-- // And this one
			}
		}
		if candidate >= 0 {
			onDelete(env.Deployed[candidate].Version)
			env.Deployed = append(env.Deployed[:candidate], env.Deployed[candidate+1:]...)
		}
	}

	// Add the new entry at the beginning of the list.
	env.Deployed = append([]HistEvent{HistEvent{versionName, time.Now()}}, env.Deployed...)

	return nil
}

// Release makes a deployed artifact the currently active one in this environment.
func (env *Env) Release(versionName string, previewers []string) error {

	// Ensure that this version of the artifact has been deployed.
	found := false
	for _, v := range env.Deployed {
		if v.Version == versionName {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("version %q not deployed", versionName)
	}

	// Ensure this version has not been disabled.
	if vers, found := env.versions[versionName]; found {
		if !vers.Enabled {
			return fmt.Errorf("version %q has been disabled", versionName)
		}
		// Mark it as having been released.
		vers.Release()
	} else {
		// This shouldn't happen, but just in case...
		return fmt.Errorf("version %q not found in environment", versionName)
	}

	// If specific hosts have been named, only they get the release as a preview.
	if len(previewers) > 0 {
		return env.releasePreview(versionName, previewers)
	}

	// The release is general, and goes out to every host.
	return env.releaseGeneral(versionName)
}

// GetCurrentVersion returns the current version for the specified host.
func (env *Env) GetCurrentVersion(hostName string) string {
	for _, previewer := range env.Previewers {
		if previewer == hostName {
			return env.Preview
		}
	}
	return env.Current
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
		env.Released = append([]HistEvent{HistEvent{versionName, time.Now()}}, env.Released...)
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

func (env *Env) isPurgable(versionName string) bool {
	if versionName == env.Current || versionName == env.Prior || versionName == env.Preview {
		return false
	}
	return true
}

func (env *Env) purgeVersion(versionName string) error {

	if versionName == env.Current || versionName == env.Prior || versionName == env.Preview {
		return fmt.Errorf("version %q in use", versionName)
	}

	deployed := make([]HistEvent, 0)
	for _, histEvent := range env.Deployed {
		if histEvent.Version != versionName {
			deployed = append(deployed, histEvent)
		}
	}
	env.Deployed = deployed

	released := make([]HistEvent, 0)
	for _, histEvent := range env.Released {
		if histEvent.Version != versionName {
			released = append(released, histEvent)
		}
	}
	env.Released = released

	return nil
}

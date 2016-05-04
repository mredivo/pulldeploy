package repo

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

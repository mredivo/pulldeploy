package repo

// Env enumerates the versions deployed to an environment, and identifies the current release.
type Env struct {
	Prev       string   `json:"prev"`       // The index into the Deployed map of a candidate version
	Current    string   `json:"current"`    // The index into the Deployed map of the current version
	Next       string   `json:"next"`       // The index into the Deployed map of a candidate version
	Deployed   []string `json:"deployed"`   // The set of versions deployed to this environment
	Released   []string `json:"released"`   // The set of versions released to this environment
	Previewers []string `json:"previewers"` // An array of hostnames eligible for the Next version
}

func newEnv() *Env {
	return &Env{Deployed: []string{}, Released: []string{}, Previewers: []string{}}
}

package repo

// Version provides version and filename information for uploaded files.
type Version struct {
	Name     string `json:"version"`  // The name of the version, expected to be similar to "1.0.0"
	Filename string `json:"filename"` // The name of the uploaded tar file for this version
	Released bool   `json:"released"` // True if this version has ever been released
	Enabled  bool   `json:"enabled"`  // False if this version has been specifically disabled; default True
}

func newVersion(versionName, fileName string, enabled bool) *Version {
	return &Version{Name: versionName, Filename: fileName, Released: false, Enabled: enabled}
}

// Enable makes a version eligible to be released (the default state).
func (vers *Version) Enable() {
	vers.Enabled = true
}

/*
Disable makes a version ineligible for release.

This is useful if a defective version has been deployed, so that it cannot be
re-released after a rollback.
*/
func (vers *Version) Disable() {
	vers.Enabled = false
}

// Release records that this version has been the current version at some point.
func (vers *Version) Release() {
	vers.Released = true
}

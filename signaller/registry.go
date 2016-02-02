package signaller

import (
	"path"
)

// RegistryInfo describes the hosts entered into the hosts registry.
type RegistryInfo struct {
	Hostname   string // The name of the server running the application
	AppVersion string // The version of the application this host is serving
}

// RegistryList is an array of RegistryInfo structures.
type registryList []RegistryInfo

// Provide Sort interface methods to allow sorting by hostname.
func (rl registryList) Len() int           { return len(rl) }
func (rl registryList) Swap(i, j int)      { rl[i], rl[j] = rl[j], rl[i] }
func (rl registryList) Less(i, j int) bool { return rl[i].Hostname < rl[j].Hostname }

// makeRegistryPath builds the Zookeeper path corresponding to the name env and app.
//   /<base>/<env>/deployments/<app>/registry/<host>
func (sgnlr *Signaller) makeRegistryPath(envName, appName, hostName string) string {
	return path.Join(sgnlr.cfg.BaseNode, envName, "deployments", appName, "registry", hostName)
}

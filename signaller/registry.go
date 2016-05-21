package signaller

import (
	"path"
	"sort"

	"github.com/samuel/go-zookeeper/zk"
)

// RegistryInfo describes the hosts entered into the hosts registry.
type RegistryInfo struct {
	Hostname   string // The name of the server running the application
	Envname    string // The name of the environment this host is tracking
	Appname    string // The name of the application this host is running
	AppVersion string // The version of the application this host is serving
}

// RegistryList is an array of RegistryInfo structures.
type registryList []RegistryInfo

// Provide Sort interface methods to allow sorting by hostname.
func (rl registryList) Len() int           { return len(rl) }
func (rl registryList) Swap(i, j int)      { rl[i], rl[j] = rl[j], rl[i] }
func (rl registryList) Less(i, j int) bool { return rl[i].Hostname < rl[j].Hostname }

/*
Registry is a registry of all hosts running Pulldeploy, with the environments
and applications they are tracking.

NOTE: The Registry stores its data in Zookeeper ephemeral nodes, so this feature
is available only when PullDeploy has access to a Zookeeper installation.
*/
type Registry struct {
	sgnlr *Signaller
}

// Register enters the name of the local machine into the hosts registry,
// along with the currently released version (requires Zookeeper).
func (hr *Registry) Register(envName, appName, hostName, version string) {
	if zkConn := hr.sgnlr.getZKConnWithLock(); zkConn != nil {
		flags := int32(zk.FlagEphemeral)
		acl := zk.WorldACL(zk.PermAll)
		registryPath := hr.makeRegistryPath(envName, appName, hostName)
		hr.sgnlr.makeParentNodes(registryPath)
		zkConn.Create(registryPath, []byte(version), flags, acl)
	}
}

// Unregister removes the name of the local machine from the hosts registry
// (requires Zookeeper).
func (hr *Registry) Unregister(envName, appName, hostName string) {
	if zkConn := hr.sgnlr.getZKConnWithLock(); zkConn != nil {
		registryPath := hr.makeRegistryPath(envName, appName, hostName)
		zkConn.Delete(registryPath, -1)
	}
}

// Hosts retrieves the information in the hosts registry for the given
// environment and application (requires Zookeeper).
func (hr *Registry) Hosts(envName, appName string) []RegistryInfo {

	var ri = make(registryList, 0)

	if zkConn := hr.sgnlr.getZKConnWithLock(); zkConn != nil {
		registryPath := hr.makeRegistryPath(envName, appName, "")
		data := make([]byte, 100)
		hosts, _, _ := zkConn.Children(registryPath)
		for _, host := range hosts {
			data, _, _ = zkConn.Get(registryPath + "/" + host)
			ri = append(ri, RegistryInfo{host, envName, appName, string(data)})
		}
	}
	sort.Sort(ri)

	return ri
}

// makeRegistryPath builds the Zookeeper path corresponding to the name env and app.
//   /<base>/<env>/deployments/<app>/registry/<host>
func (hr *Registry) makeRegistryPath(envName, appName, hostName string) string {
	return path.Join(hr.sgnlr.cfg.ZK.BaseNode, envName, "deployments", appName, "registry", hostName)
}

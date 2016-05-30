package signaller

import (
	"bytes"
	"encoding/json"
	"path"
	"sort"

	"github.com/samuel/go-zookeeper/zk"
)

// hostInfo is serialized for storage in Zookeeper.
type hostInfo struct {
	Version  string   // The version of the application this host is serving
	Deployed []string // The versions currently available on this host
}

// RegistryInfo is used to present the information in the Registry.
type RegistryInfo struct {
	Hostname   string   // The name of the server running the application
	Envname    string   // The name of the environment this host is tracking
	Appname    string   // The name of the application this host is running
	AppVersion string   // The version of the application this host is serving
	Deployed   []string // The versions currently available on this host
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

// Register enters the name of the local machine into the hosts registry, along with the
// currently released version and available deployments (requires Zookeeper).
func (hr *Registry) Register(envName, appName, hostName, version string, deployed []string) {
	if zkConn := hr.sgnlr.getZKConnWithLock(); zkConn != nil {

		hostinfo := hostInfo{version, deployed}
		data, _ := json.MarshalIndent(hostinfo, "", "    ")

		flags := int32(zk.FlagEphemeral)
		acl := zk.WorldACL(zk.PermAll)
		registryPath := hr.makeRegistryPath(envName, appName, hostName)
		hr.sgnlr.makeParentNodes(registryPath)
		if _, err := zkConn.Create(registryPath, data, flags, acl); err != nil {
			zkConn.Set(registryPath, data, -1)
		}
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
	var hostinfo hostInfo

	if zkConn := hr.sgnlr.getZKConnWithLock(); zkConn != nil {
		registryPath := hr.makeRegistryPath(envName, appName, "")
		data := make([]byte, 2048)
		hosts, _, _ := zkConn.Children(registryPath)
		for _, host := range hosts {
			data, _, _ = zkConn.Get(registryPath + "/" + host)
			decoder := json.NewDecoder(bytes.NewReader(data))
			decoder.Decode(&hostinfo)
			ri = append(ri, RegistryInfo{host, envName, appName, hostinfo.Version, hostinfo.Deployed})
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

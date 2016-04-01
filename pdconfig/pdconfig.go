package pdconfig

import (
	"fmt"
)

// ZookeeperConfig contains connection and path information.
type ZookeeperConfig struct {
	Servers  []string // Zookeeper servers: host[:port]
	BaseNode string   // The path to the base node of all Zookeeper nodes
}

// SignallerConfig contains timeouts and notification information
type SignallerConfig struct {
	PollInterval int             // Seconds between repository polls when not using Zookeeper
	PollFallback int             // Seconds between repository polls when Zookeeper is available
	ZK           ZookeeperConfig `yaml:"zookeeper"`
}

// StorageConfig contains the repository storage location, and its instantiation parameters.
type StorageConfig struct {
	Type   string            // One of the KST_* StorageType constants
	Params map[string]string // Type-specific parameters
}

// AppConfig contains the definition of each PullDeploy client application,
// loaded from /etc/pulldeploy.d/<appname>.json
type AppConfig struct {
	Description string // A short description of the application
	Secret      string // The secret used to sign the deployment package
	Directory   string // The base directory of the deployment on the app server
	User        string // The user that should own all deployed artifacts
	Group       string // The group that should own all deployed artifacts
}

// The definition of the configuration object shared throughout PullDeploy.
type PDConfig interface {
	GetSignallerConfig() *SignallerConfig
	GetStorageConfig() *StorageConfig
	GetAppConfig(appName string) (*AppConfig, error)
	GetAppList() map[string]*AppConfig
}

// GetSignallerConfig returns the polling and Zookeeper information.
func (pdcfg *pdConfig) GetSignallerConfig() *SignallerConfig {
	sc := new(SignallerConfig)
	sc.PollInterval = pdcfg.Signaller.PollInterval
	sc.PollFallback = pdcfg.Signaller.PollFallback
	sc.ZK = pdcfg.Signaller.ZK
	return sc
}

// GetStorageConfig returns the type and params for the configured storage.
func (pdcfg *pdConfig) GetStorageConfig() *StorageConfig {
	sc := new(StorageConfig)
	sc.Type = pdcfg.StorageType
	if params, found := pdcfg.Storage[pdcfg.StorageType]; found {
		sc.Params = params
	} else {
		sc.Params = make(map[string]string)
	}
	return sc
}

// GetAppConfig returns a client application configuration.
func (pdcfg *pdConfig) GetAppConfig(appName string) (*AppConfig, error) {

	var appConfig AppConfig
	if ac, found := pdcfg.appList[appName]; found {
		appConfig = *ac
		return &appConfig, nil
	} else {
		return nil, fmt.Errorf("No configuration available for application %q", appName)
	}
}

// GetAppList returns a list of the client applications.
func (pdcfg *pdConfig) GetAppList() map[string]*AppConfig {

	var appList map[string]*AppConfig = make(map[string]*AppConfig)

	for k, v := range pdcfg.appList {
		appList[k] = v
	}

	return appList
}
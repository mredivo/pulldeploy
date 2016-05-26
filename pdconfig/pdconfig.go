// Package pdconfig loads and manages all configuration.
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
	AccessMethod string            // One of the KST_* AccessMethod constants
	Params       map[string]string // Type-specific parameters
}

type sysCommand struct {
	Cmd  string
	Args []string
}

// ArtifactConfig defines the valid artifact types and how to unpack them.
type ArtifactConfig struct {
	Insecure  bool       // True if configuration was loaded from insecure file
	Extension string     // The filename extension to use for this artifact type
	Extract   sysCommand // The command used to unpack this artifact type
}

// AppConfig contains the definition of each PullDeploy client application,
// loaded from /etc/pulldeploy.d/<appname>.json
type AppConfig struct {
	Description  string // A short description of the application
	Secret       string // The secret used to sign the deployment package
	ArtifactType string // The file extension; determines unpacking method
	BaseDir      string // The base directory of the deployment on the app server
	User         string // The user that should own all deployed artifacts
	Group        string // The group that should own all deployed artifacts
	Insecure     bool   // True if configuration was loaded from insecure file
	Scripts      map[string]sysCommand
}

// The definition of the configuration object shared throughout PullDeploy.
type PDConfig interface {
	GetLogLevel() string
	GetVersionInfo() *VersionInfo
	GetSignallerConfig() *SignallerConfig
	GetStorageConfig() *StorageConfig
	GetArtifactConfig(artifactType string) (*ArtifactConfig, error)
	GetAppConfig(appName string) (*AppConfig, error)
	GetAppList() map[string]*AppConfig
	RefreshAppList() []error
}

// GetLogLevel returns the level at which to log.
func (pdcfg *pdConfig) GetLogLevel() string {
	return pdcfg.LogLevel
}

// GetVersionInfo returns the object containing baked-in version information.
func (pdcfg *pdConfig) GetVersionInfo() *VersionInfo {
	return &versionInfo
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
	sc.AccessMethod = pdcfg.AccessMethod
	if params, found := pdcfg.Storage[pdcfg.AccessMethod]; found {
		sc.Params = params
	} else {
		sc.Params = make(map[string]string)
	}
	return sc
}

// GetArtifactConfig returns a client application configuration.
func (pdcfg *pdConfig) GetArtifactConfig(artifactType string) (*ArtifactConfig, error) {

	var artifactConfig ArtifactConfig
	if ac, found := pdcfg.ArtifactTypes[artifactType]; found {
		if ac.Extract.Cmd != "" {
			// It passed validation.
			artifactConfig = ac
			return &artifactConfig, nil
		} else {
			return nil, fmt.Errorf("Incomplete configuration for artifact type %q", artifactType)
		}
	} else {
		return nil, fmt.Errorf("No configuration for artifact type %q", artifactType)
	}
}

// GetAppConfig returns a client application configuration.
func (pdcfg *pdConfig) GetAppConfig(appName string) (*AppConfig, error) {

	var appConfig AppConfig
	if ac, found := pdcfg.appList[appName]; found {
		appConfig = *ac
		return &appConfig, nil
	} else {
		return nil, fmt.Errorf("No configuration for application %q", appName)
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

// RefreshAppList re-reads the definitions of all the configured applications.
func (pdcfg *pdConfig) RefreshAppList() []error {
	var errs []error = make([]error, 0)
	if appList, appErrs := loadAppList(pdcfg.configDir); appErrs == nil {
		pdcfg.appList = appList
	} else {
		errs = appErrs
	}
	return errs
}

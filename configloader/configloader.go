// Package configloader loads and manages all application configuration.
package configloader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v2"
)

const kCONFIG_FILENAME = "pulldeploy.yaml" // The name of the main configuration file
const kCONFIG_DIR_DEV = "data/etc"         // Location of developer version of the config
const kCONFIG_DIR_PROD = "/etc"            // Location of production version of the config
const kCONFIG_APP_DIR = "pulldeploy.d"     // Subdirectory for application config files
const kCONFIG_APP_EXT = ".json"            // Filename extension for application config files

// The configuration as read in.
var sourceValues sourceConfig

type ZookeeperConfig struct {
	Servers  []string // Zookeeper servers: host[:port]
	BaseNode string   // The path to the base node of all Zookeeper nodes
}

type SignallerConfig struct {
	PollInterval int             // Seconds between repository polls when not using Zookeeper
	PollFallback int             // Seconds between repository polls when Zookeeper is available
	ZK           ZookeeperConfig `yaml:"zookeeper"`
}

type sourceConfig struct {
	configDir   string // Invisible to YAML decoder, determined at runtime
	StorageType string // One of the KST_* StorageType constants
	Storage     map[string]map[string]string
	Signaller   SignallerConfig
}

type StorageConfig struct {
	Type   string            // One of the KST_* StorageType constants
	Params map[string]string // Type-specific parameters
}

type AppConfig struct {
	Description string // A short description of the application
	Secret      string // The secret used to sign the deployment package
	Directory   string // The base directory of the deployment on the app server
	User        string // The user that should own all deployed artifacts
	Group       string // The group that should own all deployed artifacts
}

// findConfigDir enables use of a developer-specific configuration directory.
func findConfigDir() (string, error) {

	var p []string = []string{} // The directories in which we looked, for error message
	var configDir string        // The directory in which we found the configuration

	// Look for the developer's private configuration.
	if dir, err := os.Getwd(); err == nil {
		cfgdir := path.Join(dir, kCONFIG_DIR_DEV)
		cfgpath := path.Join(cfgdir, kCONFIG_FILENAME)
		p = append(p, cfgdir)
		if _, err := os.Stat(cfgpath); err == nil {
			configDir = cfgdir
		}
	}

	// If not found, look for the production configuration.
	if configDir == "" {
		cfgdir := kCONFIG_DIR_PROD
		cfgpath := path.Join(cfgdir, kCONFIG_FILENAME)
		p = append(p, cfgdir)
		if _, err := os.Stat(cfgpath); err == nil {
			configDir = cfgdir
		}
	}

	// Report an error if no configuration file was found.
	if configDir == "" {
		return "", fmt.Errorf("Unable to locate configuration file %q in path %s",
			kCONFIG_FILENAME, strings.Join(p, ":"))
	} else {
		return configDir, nil
	}
}

// LoadPulldeployConfig loads the main configuration file.
func LoadPulldeployConfig() (string, error) {

	var configFile string

	if configDir, err := findConfigDir(); err == nil {
		sourceValues.configDir = configDir
		configFile = path.Join(configDir, kCONFIG_FILENAME)
	} else {
		return "", err
	}

	text, err := ioutil.ReadFile(configFile)
	if err == nil {
		err = yaml.Unmarshal(text, &sourceValues)
	} else {
		return "", fmt.Errorf("Unable to read configuration file %q: %s", configFile, err.Error())
	}

	return configFile, err
}

// GetStorageConfig returns the type and params for the configured storage.
func GetStorageConfig() *StorageConfig {
	sc := new(StorageConfig)
	sc.Type = sourceValues.StorageType
	if params, found := sourceValues.Storage[sourceValues.StorageType]; found {
		sc.Params = params
	} else {
		sc.Params = make(map[string]string)
	}
	return sc
}

// GetAppList returns a list of the client applications.
func GetAppList() map[string]interface{} {

	var appList map[string]interface{} = make(map[string]interface{})

	if files, err := ioutil.ReadDir(path.Join(sourceValues.configDir, kCONFIG_APP_DIR)); err == nil {
		for _, file := range files {
			filename := file.Name()
			if path.Ext(filename) == ".json" {
				appName := strings.TrimSuffix(filename, kCONFIG_APP_EXT)
				if ac, err := GetAppConfig(appName); err == nil {
					appList[appName] = ac
				} else {
					appList[appName] = err
				}
			}
		}
	}

	return appList
}

// GetAppConfig returns a client application configuration.
func GetAppConfig(appName string) (*AppConfig, error) {

	appcfg := new(AppConfig)
	appcfgfile := path.Join(sourceValues.configDir, kCONFIG_APP_DIR, appName+kCONFIG_APP_EXT)

	if f, err := os.Open(appcfgfile); err == nil {
		defer f.Close()
		decoder := json.NewDecoder(f)
		if err := decoder.Decode(appcfg); err == nil {
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}

	return appcfg, nil
}

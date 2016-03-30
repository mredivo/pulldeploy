// Package configloader loads and manages all application configuration.
package configloader

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

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
	configFile    string // Invisible to YAML decoder, populated with actual file loaded
	StorageMethod string // One of the KST_* StorageType constants
	Storage       map[string]map[string]string
	Signaller     SignallerConfig
}

func LoadPulldeployConfig() (string, error) {

	var p []string = []string{}
	var configFile string
	var filename = "pulldeploy.yaml"

	// Look for the developer's private configuration.
	if dir, err := os.Getwd(); err == nil {
		p = append(p, path.Join(dir, "data/etc", filename))
		if _, err := os.Stat(p[0]); err == nil {
			configFile = p[0]
		}
	}

	// Look for the production configuration.
	if configFile == "" {
		p = append(p, path.Join("/etc", filename))
		if _, err := os.Stat(p[1]); err == nil {
			configFile = p[1]
		}
	}

	// Report an error if no configuration file was found.
	if configFile == "" {
		return "", fmt.Errorf("Unable to locate configuration file %q: Tried %v", filename, p)
	}

	text, err := ioutil.ReadFile(configFile)
	if err == nil {
		// Decode the YAML text into the environment configuration struct.
		err = yaml.Unmarshal(text, &sourceValues)
	} else {
		return "", fmt.Errorf("Unable to read configuration file %q: %s", configFile, err.Error())
	}
	sourceValues.configFile = configFile

	return configFile, err
}

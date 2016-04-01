// Package pdconfig loads and manages all application configuration.
package pdconfig

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
type pdConfig struct {
	configDir   string                // Invisible to YAML decoder, determined at runtime
	configFile  string                // Invisible to YAML decoder, determined at runtime
	appList     map[string]*AppConfig // Invisible to YAML decoder, loaded separately
	StorageType string                // One of the KST_* StorageType constants
	Storage     map[string]map[string]string
	Signaller   SignallerConfig
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

// loadAppConfig loads the configuration for a client application.
func loadAppConfig(configDir, appName string) (*AppConfig, error) {

	appcfg := new(AppConfig)
	appcfgfile := path.Join(configDir, kCONFIG_APP_DIR, appName+kCONFIG_APP_EXT)

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

// LoadPulldeployConfig loads the main configuration file and all the client apps.
func LoadPulldeployConfig() (PDConfig, []error) {

	var errs []error = make([]error, 0)
	var pdconfig *pdConfig = new(pdConfig)
	pdconfig.appList = make(map[string]*AppConfig)

	// Determine which configuration directory to use.
	if configDir, err := findConfigDir(); err == nil {
		pdconfig.configDir = configDir
		pdconfig.configFile = path.Join(configDir, kCONFIG_FILENAME)
	} else {
		errs = append(errs, err)
		return nil, errs
	}

	// Read in the YAML and decode it.
	text, err := ioutil.ReadFile(pdconfig.configFile)
	if err == nil {
		err = yaml.Unmarshal(text, &pdconfig)
	} else {
		errs = append(errs, fmt.Errorf(
			"Unable to read configuration file %q: %s",
			pdconfig.configFile, err.Error()))
		return nil, errs
	}

	// Read in all the client application definitions.
	if files, err := ioutil.ReadDir(path.Join(pdconfig.configDir, kCONFIG_APP_DIR)); err == nil {
		for _, file := range files {
			filename := file.Name()
			if path.Ext(filename) == ".json" {
				appName := strings.TrimSuffix(filename, kCONFIG_APP_EXT)
				if ac, err := loadAppConfig(pdconfig.configDir, appName); err == nil {
					pdconfig.appList[appName] = ac
				} else {
					errs = append(errs, err)
				}
			}
		}
	}

	return pdconfig, errs
}

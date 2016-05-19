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

// loadAppList reads in the definitions of all the configured applications.
func loadAppList(configDir string) (map[string]*AppConfig, []error) {

	var errs []error = make([]error, 0)
	var appList = make(map[string]*AppConfig)

	if files, err := ioutil.ReadDir(path.Join(configDir, kCONFIG_APP_DIR)); err == nil {
		for _, file := range files {
			filename := file.Name()
			if path.Ext(filename) == ".json" {
				appName := strings.TrimSuffix(filename, kCONFIG_APP_EXT)
				if ac, err := loadAppConfig(configDir, appName); err == nil {
					appList[appName] = ac
				} else {
					errs = append(errs, err)
				}
			}
		}
	}

	return appList, errs
}

// LoadPulldeployConfig loads the main configuration file and all the client apps.
func LoadPulldeployConfig(configDir string) (PDConfig, []error) {

	var errs []error = make([]error, 0)
	var pdcfg *pdConfig = new(pdConfig)

	// Determine which configuration directory to use.
	if configDir != "" {
		pdcfg.configDir = configDir
	} else {
		if configDir, err := findConfigDir(); err == nil {
			pdcfg.configDir = configDir
		} else {
			errs = append(errs, err)
			return nil, errs
		}
	}
	pdcfg.configFile = path.Join(pdcfg.configDir, kCONFIG_FILENAME)

	// Read in the YAML and decode it.
	text, err := ioutil.ReadFile(pdcfg.configFile)
	if err == nil {
		err = yaml.Unmarshal(text, &pdcfg)
	} else {
		errs = append(errs, fmt.Errorf(
			"Unable to read configuration file %q: %s",
			pdcfg.configFile, err.Error()))
		return nil, errs
	}

	// Read in all the client application definitions.
	if appList, appErrs := loadAppList(pdcfg.configDir); len(appErrs) == 0 {
		pdcfg.appList = appList
	} else {
		pdcfg.appList = make(map[string]*AppConfig)
		errs = append(errs, appErrs...)
	}

	return pdcfg, errs
}

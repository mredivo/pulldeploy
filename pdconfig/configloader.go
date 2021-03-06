package pdconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"

	"gopkg.in/yaml.v2"
)

const kCONFIG_FILENAME = "pulldeploy.yaml" // The name of the main configuration file
const kCONFIG_DIR_DEV = "data/etc"         // Location of developer version of the config
const kCONFIG_DIR_PROD = "/etc"            // Location of production version of the config
const kCONFIG_APP_DIR = "pulldeploy.d"     // Subdirectory for application config files
const kCONFIG_APP_EXT = ".yaml"            // Filename extension for application config files

// The configuration as read in.
type pdConfig struct {
	configDir     string                // Invisible to YAML decoder, determined at runtime
	configFile    string                // Invisible to YAML decoder, determined at runtime
	appList       map[string]*AppConfig // Invisible to YAML decoder, loaded separately
	LogLevel      string                // The level at which to log: debug|info|warn|error
	AccessMethod  string                // One of the KST_* AccessMethod constants
	Storage       map[string]map[string]string
	Signaller     SignallerConfig
	ArtifactTypes map[string]ArtifactConfig
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

// When running as root, configurations must writable only by root.
func isInsecure(filepath string) bool {

	isInsecure := false

	if os.Geteuid() == 0 {
		// Check whether the file is world writable.
		if fi, err := os.Stat(filepath); err == nil {
			if perm := fi.Mode().Perm(); perm&02 > 0 {
				isInsecure = true
			}
			// Check file ownership. Note: fi.Sys() is non-portable.
			if sys := fi.Sys(); sys != nil {
				if sys.(*syscall.Stat_t).Uid != 0 {
					isInsecure = true
				}
				if sys.(*syscall.Stat_t).Gid != 0 {
					isInsecure = true
				}
			}
		}
	}

	return isInsecure
}

// loadAppConfig loads the configuration for a client application.
func loadAppConfig(configDir, appName string) (*AppConfig, error) {

	appcfg := new(AppConfig)
	appcfgfile := path.Join(configDir, kCONFIG_APP_DIR, appName+kCONFIG_APP_EXT)

	// Read in the YAML and decode it.
	text, err := ioutil.ReadFile(appcfgfile)
	if err == nil {
		if err = yaml.Unmarshal(text, &appcfg); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	// When running as root, configurations must be secure.
	appcfg.Insecure = isInsecure(appcfgfile)

	return appcfg, nil
}

// loadAppList reads in the definitions of all the configured applications.
func loadAppList(configDir string) (map[string]*AppConfig, []error) {

	var errs []error = make([]error, 0)
	var appList = make(map[string]*AppConfig)

	if files, err := ioutil.ReadDir(path.Join(configDir, kCONFIG_APP_DIR)); err == nil {
		for _, file := range files {
			filename := file.Name()
			if path.Ext(filename) == kCONFIG_APP_EXT {
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

	// When running as root, configurations must be secure.
	isInsecure := isInsecure(pdcfg.configFile)

	// Validate the system-specific shell commands.
	var allOK = true
	for atype, acfg := range pdcfg.ArtifactTypes {
		if acfg.Extract.Cmd != "" {
			if _, err := os.Stat(acfg.Extract.Cmd); os.IsNotExist(err) {
				errs = append(errs, fmt.Errorf(
					"ArtifactType %q extract command error: %s",
					atype, err.Error()))
				allOK = false
			}
		}
		if isInsecure {
			acfg.Insecure = isInsecure
			pdcfg.ArtifactTypes[atype] = acfg
		}
	}
	if !allOK {
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

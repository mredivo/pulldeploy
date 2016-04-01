package command

import (
	"flag"
	"fmt"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// pulldeploy enable -app=<app> -version=<version>
type Enable struct {
	pdcfg      pdconfig.PDConfig
	appName    string
	appVersion string
}

func (cmd *Enable) CheckArgs(pdcfg pdconfig.PDConfig, osArgs []string) bool {

	cmd.pdcfg = pdcfg

	var appName, appVersion string

	cmdFlags := flag.NewFlagSet("enable", flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&appVersion, "version", "", "version of the application being enabled")
	cmdFlags.Parse(osArgs)

	isValid := true

	if appName == "" {
		fmt.Println("app is a mandatory argument")
		isValid = false
	} else {
		cmd.appName = appName
	}

	if appVersion == "" {
		fmt.Println("version is a mandatory argument")
		isValid = false
	} else {
		cmd.appVersion = appVersion
	}

	return isValid
}

func (cmd *Enable) Exec() {
	fmt.Printf("enable(%s, %s)\n", cmd.appName, cmd.appVersion)
}

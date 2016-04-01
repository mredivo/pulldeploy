package command

import (
	"flag"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// pulldeploy disable -app=<app> -version=<version>
type Disable struct {
	el         *ErrorList
	pdcfg      pdconfig.PDConfig
	appName    string
	appVersion string
}

func (cmd *Disable) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

	var appName, appVersion string
	cmd.el = NewErrorList(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet("disable", flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&appVersion, "version", "", "version of the application being disabled")
	cmdFlags.Parse(osArgs)

	if appName == "" {
		cmd.el.Errorf("app is a mandatory argument")
	} else {
		cmd.appName = appName
	}

	if appVersion == "" {
		cmd.el.Errorf("version is a mandatory argument")
	} else {
		cmd.appVersion = appVersion
	}

	return cmd.el
}

func (cmd *Disable) Exec() *ErrorList {
	placeHolder("disable(%s, %s)\n", cmd.appName, cmd.appVersion)
	return cmd.el
}

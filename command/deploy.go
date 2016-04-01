package command

import (
	"flag"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// pulldeploy deploy -app=<app> -version=<version> -env=<env>
type Deploy struct {
	el         *ErrorList
	pdcfg      pdconfig.PDConfig
	appName    string
	appVersion string
	envName    string
}

func (cmd *Deploy) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

	var appName, appVersion, envName string
	cmd.el = NewErrorList(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet("deploy", flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&appVersion, "version", "", "version of the application to be deployed")
	cmdFlags.StringVar(&envName, "env", "", "environment to which to deploy")
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

	if envName == "" {
		cmd.el.Errorf("env is a mandatory argument")
	} else {
		cmd.envName = envName
	}

	return cmd.el
}

func (cmd *Deploy) Exec() {
	placeHolder("deploy(%s, %s, %s)\n", cmd.appName, cmd.appVersion, cmd.envName)
}

package command

import (
	"flag"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// pulldeploy listhosts -app=<app> -env=<env>
type Listhosts struct {
	el      *ErrorList
	pdcfg   pdconfig.PDConfig
	appName string
	envName string
}

func (cmd *Listhosts) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

	var appName, envName string
	cmd.el = NewErrorList(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet("listhosts", flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&envName, "env", "", "environment in which to release")
	cmdFlags.Parse(osArgs)

	if appName == "" {
		cmd.el.Errorf("app is a mandatory argument")
	} else {
		cmd.appName = appName
	}

	if envName == "" {
		cmd.el.Errorf("env is a mandatory argument")
	} else {
		cmd.envName = envName
	}

	return cmd.el
}

func (cmd *Listhosts) Exec() {
	placeHolder("listhosts(%s, %s)\n", cmd.appName, cmd.envName)
}

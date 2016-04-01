package command

import (
	"flag"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// pulldeploy addenv -app=<app> envname [envname envname ...]
type Addenv struct {
	el       *ErrorList
	pdcfg    pdconfig.PDConfig
	appName  string
	envNames []string
}

func (cmd *Addenv) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

	var appName string
	cmd.el = NewErrorList(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet("addenv", flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.Parse(osArgs)

	if appName == "" {
		cmd.el.Errorf("app is a mandatory argument")
	} else {
		cmd.appName = appName
	}

	if len(cmdFlags.Args()) < 1 {
		cmd.el.Errorf("at least 1 environment name must be specified")
	} else {
		cmd.envNames = cmdFlags.Args()
	}

	return cmd.el
}

func (cmd *Addenv) Exec() *ErrorList {
	placeHolder("addenv(%s, %v)\n", cmd.appName, cmd.envNames)
	return cmd.el
}

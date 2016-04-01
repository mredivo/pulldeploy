package command

import (
	"flag"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// pulldeploy set -app=<app> [-keep=n]
type Set struct {
	el      *ErrorList
	pdcfg   pdconfig.PDConfig
	appName string
	keep    int
}

func (cmd *Set) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

	var appName string
	var keep int
	cmd.el = NewErrorList(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet("set", flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application whose repository to update")
	cmdFlags.IntVar(&keep, "keep", 5, "the number of versions of app to keep in the repository")
	cmdFlags.Parse(osArgs)

	if appName == "" {
		cmd.el.Errorf("app is a mandatory argument")
	} else {
		cmd.appName = appName
	}

	if keep < 2 {
		cmd.el.Errorf("keep must be at least 2")
	} else {
		cmd.keep = keep
	}

	return cmd.el
}

func (cmd *Set) Exec() {
	placeHolder("set(%s, %d)\n", cmd.appName, cmd.keep)
}

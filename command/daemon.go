package command

import (
	"flag"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// pulldeploy daemon ...
type Daemon struct {
	el      *ErrorList
	pdcfg   pdconfig.PDConfig
	envName string
}

func (cmd *Daemon) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

	var envName string
	cmd.el = NewErrorList(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet("daemon", flag.ExitOnError)
	cmdFlags.StringVar(&envName, "env", "", "environment to be monitored")
	cmdFlags.Parse(osArgs)

	if envName == "" {
		cmd.el.Errorf("env is a mandatory argument")
	} else {
		cmd.envName = envName
	}

	return cmd.el
}

func (cmd *Daemon) Exec() {
	placeHolder("daemon(%s)\n", cmd.envName)
}

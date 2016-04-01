package command

import (
	"flag"
	"fmt"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// pulldeploy daemon ...
type Daemon struct {
	pdcfg   pdconfig.PDConfig
	envName string
}

func (cmd *Daemon) CheckArgs(pdcfg pdconfig.PDConfig, osArgs []string) bool {

	cmd.pdcfg = pdcfg

	var envName string

	cmdFlags := flag.NewFlagSet("daemon", flag.ExitOnError)
	cmdFlags.StringVar(&envName, "env", "", "environment to be monitored")
	cmdFlags.Parse(osArgs)

	isValid := true

	if envName == "" {
		fmt.Println("env is a mandatory argument")
		isValid = false
	} else {
		cmd.envName = envName
	}

	return isValid
}

func (cmd *Daemon) Exec() {
	fmt.Printf("daemon(%s)\n", cmd.envName)
}

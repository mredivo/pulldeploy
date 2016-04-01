package command

import (
	"flag"
	"fmt"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// pulldeploy rmenv -app=<app> envname [envname envname ...]
type Rmenv struct {
	pdcfg    pdconfig.PDConfig
	appName  string
	envNames []string
}

func (cmd *Rmenv) CheckArgs(pdcfg pdconfig.PDConfig, osArgs []string) bool {

	cmd.pdcfg = pdcfg

	var appName string

	cmdFlags := flag.NewFlagSet("rmenv", flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.Parse(osArgs)

	isValid := true

	if appName == "" {
		fmt.Println("app is a mandatory argument")
		isValid = false
	} else {
		cmd.appName = appName
	}

	if len(cmdFlags.Args()) < 1 {
		fmt.Println("at least 1 environment name must be specified")
		isValid = false
	} else {
		cmd.envNames = cmdFlags.Args()
	}

	return isValid
}

func (cmd *Rmenv) Exec() {
	fmt.Printf("rmenv(%s, %v)\n", cmd.appName, cmd.envNames)
}

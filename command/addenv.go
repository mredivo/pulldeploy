package command

import (
	"flag"
	"fmt"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// pulldeploy addenv -app=<app> envname [envname envname ...]
type Addenv struct {
	pdcfg    pdconfig.PDConfig
	appName  string
	envNames []string
}

func (cmd *Addenv) CheckArgs(pdcfg pdconfig.PDConfig, osArgs []string) bool {

	cmd.pdcfg = pdcfg

	var appName string

	cmdFlags := flag.NewFlagSet("addenv", flag.ExitOnError)
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

func (cmd *Addenv) Exec() {
	fmt.Printf("addenv(%s, %v)\n", cmd.appName, cmd.envNames)
}

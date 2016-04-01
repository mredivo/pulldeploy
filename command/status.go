package command

import (
	"flag"
	"fmt"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// pulldeploy status -app=<app>
type Status struct {
	pdcfg   pdconfig.PDConfig
	appName string
}

func (cmd *Status) CheckArgs(pdcfg pdconfig.PDConfig, osArgs []string) bool {

	cmd.pdcfg = pdcfg

	var appName string

	cmdFlags := flag.NewFlagSet("status", flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.Parse(osArgs)

	isValid := true

	if appName == "" {
		fmt.Println("app is a mandatory argument")
		isValid = false
	} else {
		cmd.appName = appName
	}

	return isValid
}

func (cmd *Status) Exec() {
	fmt.Printf("status(%s)\n", cmd.appName)
}

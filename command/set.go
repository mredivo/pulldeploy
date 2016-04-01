package command

import (
	"flag"
	"fmt"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// pulldeploy set -app=<app> [-keep=n]
type Set struct {
	pdcfg   pdconfig.PDConfig
	appName string
	keep    int
}

func (cmd *Set) CheckArgs(pdcfg pdconfig.PDConfig, osArgs []string) bool {

	cmd.pdcfg = pdcfg

	var appName string
	var keep int

	cmdFlags := flag.NewFlagSet("set", flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application whose repository to update")
	cmdFlags.IntVar(&keep, "keep", 5, "the number of versions of app to keep in the repository")
	cmdFlags.Parse(osArgs)

	isValid := true

	if appName == "" {
		fmt.Println("app is a mandatory argument")
		isValid = false
	} else {
		cmd.appName = appName
	}

	if keep < 2 {
		fmt.Println("keep must be at least 2")
		isValid = false
	} else {
		cmd.keep = keep
	}

	return isValid
}

func (cmd *Set) Exec() {
	fmt.Printf("set(%s, %d)\n", cmd.appName, cmd.keep)
}

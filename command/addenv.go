package command

import (
	"flag"
	"fmt"
)

// pulldeploy addenv -app=<app> envname [envname envname ...]
type Addenv struct {
	appName  string
	envNames []string
}

func (cmd *Addenv) CheckArgs(osArgs []string) bool {

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

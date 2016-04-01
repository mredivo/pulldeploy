package command

import (
	"flag"
	"fmt"
)

// pulldeploy status -app=<app>
type Status struct {
	appName string
}

func (cmd *Status) CheckArgs(osArgs []string) bool {

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

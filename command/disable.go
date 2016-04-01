package command

import (
	"flag"
	"fmt"
)

// pulldeploy disable -app=<app> -version=<version>
type Disable struct {
	appName    string
	appVersion string
}

func (cmd *Disable) CheckArgs(osArgs []string) bool {

	var appName, appVersion string

	cmdFlags := flag.NewFlagSet("disable", flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&appVersion, "version", "", "version of the application being disabled")
	cmdFlags.Parse(osArgs)

	isValid := true

	if appName == "" {
		fmt.Println("app is a mandatory argument")
		isValid = false
	} else {
		cmd.appName = appName
	}

	if appVersion == "" {
		fmt.Println("version is a mandatory argument")
		isValid = false
	} else {
		cmd.appVersion = appVersion
	}

	return isValid
}

func (cmd *Disable) Exec() {
	fmt.Printf("disable(%s, %s)\n", cmd.appName, cmd.appVersion)
}

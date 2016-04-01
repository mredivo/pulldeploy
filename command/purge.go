package command

import (
	"flag"
	"fmt"
)

// pulldeploy purge -app=<app> -version=<version>
type Purge struct {
	appName    string
	appVersion string
}

func (cmd *Purge) CheckArgs(osArgs []string) bool {

	var appName, appVersion string

	cmdFlags := flag.NewFlagSet("purge", flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&appVersion, "version", "", "version of the application being purged")
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

func (cmd *Purge) Exec() {
	fmt.Printf("purge(%s, %s)\n", cmd.appName, cmd.appVersion)
}

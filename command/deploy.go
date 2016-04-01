package command

import (
	"flag"
	"fmt"
)

// pulldeploy deploy -app=<app> -version=<version> -env=<env>
type Deploy struct {
	appName    string
	appVersion string
	envName    string
}

func (cmd *Deploy) CheckArgs(osArgs []string) bool {

	var appName, appVersion, envName string

	cmdFlags := flag.NewFlagSet("deploy", flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&appVersion, "version", "", "version of the application to be deployed")
	cmdFlags.StringVar(&envName, "env", "", "environment to which to deploy")
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

	if envName == "" {
		fmt.Println("env is a mandatory argument")
		isValid = false
	} else {
		cmd.envName = envName
	}

	return isValid
}

func (cmd *Deploy) Exec() {
	fmt.Printf("deploy(%s, %s, %s)\n", cmd.appName, cmd.appVersion, cmd.envName)
}

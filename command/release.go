package command

import (
	"flag"
	"fmt"
)

// pulldeploy release -app=<app> -version=<version> -env=<env> [host1, host2, ...]
type Release struct {
	appName    string
	appVersion string
	envName    string
	hosts      []string
}

func (cmd *Release) CheckArgs(osArgs []string) bool {

	var appName, appVersion, envName string

	cmdFlags := flag.NewFlagSet("release", flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&appVersion, "version", "", "version of the application to be released")
	cmdFlags.StringVar(&envName, "env", "", "environment in which to release")
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

	cmd.hosts = cmdFlags.Args()

	return isValid
}

func (cmd *Release) Exec() {
	fmt.Printf("deploy(%s, %s, %s, %v)\n", cmd.appName, cmd.appVersion, cmd.envName, cmd.hosts)
}

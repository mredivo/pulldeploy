package command

import (
	"fmt"
)

// pulldeploy release -app=<app> -version=<version> -env=<env> [host1, host2, ...]
type Release struct {
	appName    string
	appVersion string
	envName    string
	hosts      []string
}

func (cmd *Release) CheckArgs(appName, appVersion, envName string, hosts []string) bool {
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
	cmd.hosts = hosts
	return isValid
}

func (cmd *Release) Exec() {
	fmt.Printf("deploy(%s, %s, %s, %v)\n", cmd.appName, cmd.appVersion, cmd.envName, cmd.hosts)
}

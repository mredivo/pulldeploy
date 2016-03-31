package command

import (
	"fmt"
)

// pulldeploy deploy -app=<app> -version=<version> -env=<env>
type Deploy struct {
	appName    string
	appVersion string
	envName    string
}

func (cmd *Deploy) CheckArgs(appName, appVersion, envName string) bool {
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

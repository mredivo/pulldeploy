package command

import (
	"fmt"
)

// pulldeploy enable -app=<app> -version=<version>
type Enable struct {
	appName    string
	appVersion string
}

func (cmd *Enable) CheckArgs(appName, appVersion string) bool {
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

func (cmd *Enable) Exec() {
	fmt.Printf("enable(%s, %s)\n", cmd.appName, cmd.appVersion)
}

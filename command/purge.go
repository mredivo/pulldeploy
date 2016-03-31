package command

import (
	"fmt"
)

// pulldeploy purge -app=<app> -version=<version>
type Purge struct {
	appName    string
	appVersion string
}

func (cmd *Purge) CheckArgs(appName, appVersion string) bool {
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

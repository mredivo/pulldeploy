package command

import (
	"fmt"
)

// pulldeploy disable -app=<app> -version=<version>
type Disable struct {
	appName    string
	appVersion string
}

func (cmd *Disable) CheckArgs(appName, appVersion string) bool {
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

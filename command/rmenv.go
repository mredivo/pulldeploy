package command

import (
	"fmt"
)

// pulldeploy rmenv -app=<app> envname [envname envname ...]
type Rmenv struct {
	appName  string
	envNames []string
}

func (cmd *Rmenv) CheckArgs(appName string, envNames []string) bool {
	isValid := true
	if appName == "" {
		fmt.Println("app is a mandatory argument")
		isValid = false
	} else {
		cmd.appName = appName
	}
	if len(envNames) < 1 {
		fmt.Println("at least 1 environment name must be specified")
		isValid = false
	} else {
		cmd.envNames = envNames
	}
	return isValid
}

func (cmd *Rmenv) Exec() {
	fmt.Printf("rmenv(%s, %v)\n", cmd.appName, cmd.envNames)
}

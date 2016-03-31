package command

import (
	"fmt"
)

// pulldeploy listhosts -app=<app> -env=<env>
type Listhosts struct {
	appName string
	envName string
}

func (cmd *Listhosts) CheckArgs(appName, envName string) bool {
	isValid := true
	if appName == "" {
		fmt.Println("app is a mandatory argument")
		isValid = false
	} else {
		cmd.appName = appName
	}
	if envName == "" {
		fmt.Println("env is a mandatory argument")
		isValid = false
	} else {
		cmd.envName = envName
	}
	return isValid
}

func (cmd *Listhosts) Exec() {
	fmt.Printf("listhosts(%s, %s)\n", cmd.appName, cmd.envName)
}

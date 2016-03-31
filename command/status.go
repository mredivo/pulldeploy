package command

import (
	"fmt"
)

// pulldeploy status -app=<app>
type Status struct {
	appName string
}

func (cmd *Status) CheckArgs(appName string) bool {
	isValid := true
	if appName == "" {
		fmt.Println("app is a mandatory argument")
		isValid = false
	} else {
		cmd.appName = appName
	}
	return isValid
}

func (cmd *Status) Exec() {
	fmt.Printf("status(%s)\n", cmd.appName)
}

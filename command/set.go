package command

import (
	"fmt"
)

// pulldeploy set -app=<app> [-keep=n]
type Set struct {
	appName string
	keep    int
}

func (cmd *Set) CheckArgs(appName string, keep int) bool {
	isValid := true
	if appName == "" {
		fmt.Println("app is a mandatory argument")
		isValid = false
	} else {
		cmd.appName = appName
	}
	if keep < 2 {
		fmt.Println("keep must be at least 2")
		isValid = false
	} else {
		cmd.keep = keep
	}
	return isValid
}

func (cmd *Set) Exec() {
	fmt.Printf("set(%s, %d)\n", cmd.appName, cmd.keep)
}

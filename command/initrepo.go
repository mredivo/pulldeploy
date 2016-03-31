package command

import (
	"fmt"
)

// pulldeploy initrepo -app=<app> [-keep=n]
type Initrepo struct {
	appName string
	keep    int
}

func (cmd *Initrepo) CheckArgs(appName string, keep int) bool {
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

func (cmd *Initrepo) Exec() {
	fmt.Printf("initrepo(%s, %d)\n", cmd.appName, cmd.keep)
}

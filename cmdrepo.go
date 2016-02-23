package main

import (
	"fmt"
)

// pulldeploy initrepo -app=<app> [-keep=n]
type cmdInitrepo struct {
	appName string
	keep    int
}

func (cmd *cmdInitrepo) checkArgs(appName string, keep int) bool {
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

func (cmd *cmdInitrepo) exec() {
	fmt.Printf("initrepo(%s, %d)\n", cmd.appName, cmd.keep)
}

// pulldeploy set -app=<app> [-keep=n]
type cmdSet struct {
	appName string
	keep    int
}

func (cmd *cmdSet) checkArgs(appName string, keep int) bool {
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

func (cmd *cmdSet) exec() {
	fmt.Printf("set(%s, %d)\n", cmd.appName, cmd.keep)
}

// pulldeploy addenv -app=<app> envname [envname envname ...]
type cmdAddenv struct {
	appName  string
	envNames []string
}

func (cmd *cmdAddenv) checkArgs(appName string, envNames []string) bool {
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

func (cmd *cmdAddenv) exec() {
	fmt.Printf("addenv(%s, %v)\n", cmd.appName, cmd.envNames)
}

// pulldeploy rmenv -app=<app> envname [envname envname ...]
type cmdRmenv struct {
	appName  string
	envNames []string
}

func (cmd *cmdRmenv) checkArgs(appName string, envNames []string) bool {
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

func (cmd *cmdRmenv) exec() {
	fmt.Printf("addenv(%s, %v)\n", cmd.appName, cmd.envNames)
}

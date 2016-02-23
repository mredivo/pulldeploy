package main

import (
	"fmt"
)

// pulldeploy list
type cmdList struct {
}

func (cmd *cmdList) checkArgs() bool {
	return true
}

func (cmd *cmdList) exec() {
	fmt.Printf("list()\n")
}

// pulldeploy status -app=<app>
type cmdStatus struct {
	appName string
}

func (cmd *cmdStatus) checkArgs(appName string) bool {
	isValid := true
	if appName == "" {
		fmt.Println("app is a mandatory argument")
		isValid = false
	} else {
		cmd.appName = appName
	}
	return isValid
}

func (cmd *cmdStatus) exec() {
	fmt.Printf("status(%s)\n", cmd.appName)
}

// pulldeploy listhosts -app=<app> -env=<env>
type cmdListhosts struct {
	appName string
	envName string
}

func (cmd *cmdListhosts) checkArgs(appName, envName string) bool {
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

func (cmd *cmdListhosts) exec() {
	fmt.Printf("listhosts(%s, %s)\n", cmd.appName, cmd.envName)
}

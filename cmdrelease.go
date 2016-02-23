package main

import (
	"fmt"
)

// pulldeploy upload -app=<app> -version=<version> [-disabled] <file>
type cmdUpload struct {
	appName    string
	appVersion string
	disabled   bool
	filename   string
}

func (cmd *cmdUpload) checkArgs(appName, appVersion string, disabled bool, args []string) bool {
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
	cmd.disabled = disabled
	if len(args) < 1 {
		fmt.Println("filename is a mandatory argument")
		isValid = false
	} else if len(args) > 1 {
		fmt.Println("only one filename may be specified")
		isValid = false
	} else {
		cmd.filename = args[0]
	}
	return isValid
}

func (cmd *cmdUpload) exec() {
	fmt.Printf("upload(%s, %s, %v, %s)\n", cmd.appName, cmd.appVersion, cmd.disabled, cmd.filename)
}

// pulldeploy enable -app=<app> -version=<version>
type cmdEnable struct {
	appName    string
	appVersion string
}

func (cmd *cmdEnable) checkArgs(appName, appVersion string) bool {
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

func (cmd *cmdEnable) exec() {
	fmt.Printf("enable(%s, %s)\n", cmd.appName, cmd.appVersion)
}

// pulldeploy disable -app=<app> -version=<version>
type cmdDisable struct {
	appName    string
	appVersion string
}

func (cmd *cmdDisable) checkArgs(appName, appVersion string) bool {
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

func (cmd *cmdDisable) exec() {
	fmt.Printf("disable(%s, %s)\n", cmd.appName, cmd.appVersion)
}

// pulldeploy purge -app=<app> -version=<version>
type cmdPurge struct {
	appName    string
	appVersion string
}

func (cmd *cmdPurge) checkArgs(appName, appVersion string) bool {
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

func (cmd *cmdPurge) exec() {
	fmt.Printf("purge(%s, %s)\n", cmd.appName, cmd.appVersion)
}

// pulldeploy deploy -app=<app> -version=<version> -env=<env>
type cmdDeploy struct {
	appName    string
	appVersion string
	envName    string
}

func (cmd *cmdDeploy) checkArgs(appName, appVersion, envName string) bool {
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
	if envName == "" {
		fmt.Println("env is a mandatory argument")
		isValid = false
	} else {
		cmd.envName = envName
	}
	return isValid
}

func (cmd *cmdDeploy) exec() {
	fmt.Printf("deploy(%s, %s, %s)\n", cmd.appName, cmd.appVersion, cmd.envName)
}

// pulldeploy release -app=<app> -version=<version> -env=<env> [host1, host2, ...]
type cmdRelease struct {
	appName    string
	appVersion string
	envName    string
	hosts      []string
}

func (cmd *cmdRelease) checkArgs(appName, appVersion, envName string, hosts []string) bool {
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
	if envName == "" {
		fmt.Println("env is a mandatory argument")
		isValid = false
	} else {
		cmd.envName = envName
	}
	cmd.hosts = hosts
	return isValid
}

func (cmd *cmdRelease) exec() {
	fmt.Printf("deploy(%s, %s, %s, %v)\n", cmd.appName, cmd.appVersion, cmd.envName, cmd.hosts)
}

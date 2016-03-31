package command

import (
	"fmt"
)

// pulldeploy upload -app=<app> -version=<version> [-disabled] <file>
type Upload struct {
	appName    string
	appVersion string
	disabled   bool
	filename   string
}

func (cmd *Upload) CheckArgs(appName, appVersion string, disabled bool, args []string) bool {
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

func (cmd *Upload) Exec() {
	fmt.Printf("upload(%s, %s, %v, %s)\n", cmd.appName, cmd.appVersion, cmd.disabled, cmd.filename)
}

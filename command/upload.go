package command

import (
	"flag"
	"fmt"
)

// pulldeploy upload -app=<app> -version=<version> [-disabled] <file>
type Upload struct {
	appName    string
	appVersion string
	disabled   bool
	filename   string
}

func (cmd *Upload) CheckArgs(osArgs []string) bool {

	var appName, appVersion string
	var disabled bool

	cmdFlags := flag.NewFlagSet("upload", flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&appVersion, "version", "", "version of the application being uploaded")
	cmdFlags.BoolVar(&disabled, "disabled", false, "upload in disabled state")
	cmdFlags.Parse(osArgs)

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

	if len(cmdFlags.Args()) < 1 {
		fmt.Println("filename is a mandatory argument")
		isValid = false
	} else if len(cmdFlags.Args()) > 1 {
		fmt.Println("only one filename may be specified")
		isValid = false
	} else {
		cmd.filename = cmdFlags.Args()[0]
	}

	return isValid
}

func (cmd *Upload) Exec() {
	fmt.Printf("upload(%s, %s, %v, %s)\n", cmd.appName, cmd.appVersion, cmd.disabled, cmd.filename)
}

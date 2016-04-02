package command

import (
	"flag"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// pulldeploy upload -app=<app> -version=<version> [-disabled] <file>
type Upload struct {
	el         *ErrorList
	pdcfg      pdconfig.PDConfig
	appName    string
	appVersion string
	disabled   bool
	filename   string
}

func (cmd *Upload) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

	var appName, appVersion string
	var disabled bool
	cmd.el = NewErrorList(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&appVersion, "version", "", "version of the application being uploaded")
	cmdFlags.BoolVar(&disabled, "disabled", false, "upload in disabled state")
	cmdFlags.Parse(osArgs)

	if appName == "" {
		cmd.el.Errorf("app is a mandatory argument")
	} else {
		cmd.appName = appName
	}

	if appVersion == "" {
		cmd.el.Errorf("version is a mandatory argument")
	} else {
		cmd.appVersion = appVersion
	}

	cmd.disabled = disabled

	if len(cmdFlags.Args()) < 1 {
		cmd.el.Errorf("filename is a mandatory argument")
	} else if len(cmdFlags.Args()) > 1 {
		cmd.el.Errorf("only one filename may be specified")
	} else {
		cmd.filename = cmdFlags.Args()[0]
	}

	return cmd.el
}

func (cmd *Upload) Exec() *ErrorList {
	placeHolder("upload(%s, %s, %v, %s)\n", cmd.appName, cmd.appVersion, cmd.disabled, cmd.filename)
	return cmd.el
}

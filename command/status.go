package command

import (
	"flag"
	"fmt"

	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/storage"
)

// pulldeploy status -app=<app>
type Status struct {
	el      *ErrorList
	pdcfg   pdconfig.PDConfig
	appName string
}

func (cmd *Status) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

	var appName string
	cmd.el = NewErrorList(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.Parse(osArgs)

	if appName == "" {
		cmd.el.Errorf("app is a mandatory argument")
	} else {
		cmd.appName = appName
	}

	return cmd.el
}

func (cmd *Status) Exec() *ErrorList {

	// Ensure the app definition exists.
	if _, err := cmd.pdcfg.GetAppConfig(cmd.appName); err != nil {
		cmd.el.Append(err)
		return cmd.el
	}

	// Get access to the repo storage.
	stgcfg := cmd.pdcfg.GetStorageConfig()
	stg, err := storage.NewStorage(stgcfg.Type, stgcfg.Params)
	if err != nil {
		cmd.el.Append(err)
		return cmd.el
	}

	// Retrieve the repository index.
	if ri, err := getRepoIndex(stg, cmd.appName); err == nil {

		// Print a summary of the state of the application.
		// TODO: replace this simple JSON dump
		if text, err := ri.ToJSON(); err == nil {
			fmt.Println(string(text))
		} else {
			cmd.el.Append(err)
			return cmd.el
		}
	} else {
		cmd.el.Append(err)
	}

	return cmd.el
}

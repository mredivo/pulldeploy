package command

import (
	"flag"

	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/storage"
)

// pulldeploy set -app=<app> [-keep=n]
type Set struct {
	el      *ErrorList
	pdcfg   pdconfig.PDConfig
	appName string
	keep    int
}

func (cmd *Set) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

	var appName string
	var keep int
	cmd.el = NewErrorList(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application whose repository to update")
	cmdFlags.IntVar(&keep, "keep", 5, "the number of versions of app to keep in the repository")
	cmdFlags.Parse(osArgs)

	if appName == "" {
		cmd.el.Errorf("app is a mandatory argument")
	} else {
		cmd.appName = appName
	}

	if keep < 2 {
		cmd.el.Errorf("keep must be at least 2")
	} else {
		cmd.keep = keep
	}

	return cmd.el
}

func (cmd *Set) Exec() *ErrorList {

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

	// Retrieve the repository index and update it.
	if ri, err := getRepoIndex(stg, cmd.appName); err == nil {
		ri.Keep = cmd.keep
		if err := setRepoIndex(stg, ri); err != nil {
			cmd.el.Append(err)
		}
	} else {
		cmd.el.Append(err)
	}

	return cmd.el
}

package command

import (
	"flag"

	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/repo"
	"github.com/mredivo/pulldeploy/storage"
)

// pulldeploy initrepo -app=<app> [-keep=n]
type Initrepo struct {
	el      *ErrorList
	pdcfg   pdconfig.PDConfig
	appName string
	keep    int
}

func (cmd *Initrepo) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

	var appName string
	var keep int
	cmd.el = NewErrorList(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application to create in the repository")
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

func (cmd *Initrepo) Exec() *ErrorList {

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

	// Do not overwrite an existing index.
	if _, err := getRepoIndex(stg, cmd.appName); err == nil {
		cmd.el.Errorf("repository already initialized, no action taken")
		return cmd.el
	}

	// Initialize the index and store it.
	ri := repo.NewRepoIndex(cmd.appName)
	ri.Keep = cmd.keep
	if err := setRepoIndex(stg, ri); err != nil {
		cmd.el.Append(err)
	}

	return cmd.el
}

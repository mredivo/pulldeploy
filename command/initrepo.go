package command

import (
	"flag"

	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/repo"
	"github.com/mredivo/pulldeploy/storage"
)

// pulldeploy initrepo -app=<app>
type Initrepo struct {
	result  *Result
	pdcfg   pdconfig.PDConfig
	appName string
}

func (cmd *Initrepo) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *Result {

	var appName string
	cmd.result = NewResult(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application to create in the repository")
	cmdFlags.Parse(osArgs)

	if appName == "" {
		cmd.result.Errorf("app is a mandatory argument")
	} else {
		cmd.appName = appName
	}

	return cmd.result
}

func (cmd *Initrepo) Exec() *Result {

	// Ensure the app definition exists.
	if _, err := cmd.pdcfg.GetAppConfig(cmd.appName); err != nil {
		cmd.result.AppendError(err)
		return cmd.result
	}

	// Get access to the repo storage.
	stgcfg := cmd.pdcfg.GetStorageConfig()
	stg, err := storage.New(storage.AccessMethod(stgcfg.AccessMethod), stgcfg.Params)
	if err != nil {
		cmd.result.AppendError(err)
		return cmd.result
	}

	// Do not overwrite an existing index.
	if _, err := getRepoIndex(stg, cmd.appName); err == nil {
		cmd.result.Errorf("repository already initialized, no action taken")
		return cmd.result
	}

	// Initialize the index and store it.
	ri := repo.NewIndex(cmd.appName)
	if err := setRepoIndex(stg, ri); err != nil {
		cmd.result.AppendError(err)
	}

	return cmd.result
}

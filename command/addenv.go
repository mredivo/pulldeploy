package command

import (
	"flag"

	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/storage"
)

// pulldeploy addenv -app=<app> envname [envname envname ...]
type Addenv struct {
	result   *Result
	pdcfg    pdconfig.PDConfig
	appName  string
	envNames []string
}

func (cmd *Addenv) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *Result {

	var appName string
	cmd.result = NewResult(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.Parse(osArgs)

	if appName == "" {
		cmd.result.Errorf("app is a mandatory argument")
	} else {
		cmd.appName = appName
	}

	if len(cmdFlags.Args()) < 1 {
		cmd.result.Errorf("at least 1 environment name must be specified")
	} else {
		cmd.envNames = cmdFlags.Args()
	}

	return cmd.result
}

func (cmd *Addenv) Exec() *Result {

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

	// Retrieve the repository index and update it.
	if ri, err := getRepoIndex(stg, cmd.appName); err == nil {

		successCount := 0
		for _, envName := range cmd.envNames {
			if err := ri.AddEnv(envName); err != nil {
				cmd.result.AppendError(err)
			} else {
				successCount++
			}
		}
		if successCount == 0 {
			return cmd.result
		}

		if err := setRepoIndex(stg, ri); err != nil {
			cmd.result.AppendError(err)
		}
	} else {
		cmd.result.AppendError(err)
	}

	return cmd.result
}

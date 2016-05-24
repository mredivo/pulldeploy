package command

import (
	"flag"

	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/storage"
)

// pulldeploy set -app=<app> [-keep=n]
type Set struct {
	result  *Result
	pdcfg   pdconfig.PDConfig
	appName string
	envName string
	keep    int
}

func (cmd *Set) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *Result {

	var appName, envName string
	var keep int
	cmd.result = NewResult(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application whose repository to update")
	cmdFlags.StringVar(&envName, "env", "", "environment to update")
	cmdFlags.IntVar(&keep, "keep", 5, "the number of versions of app to keep in the environment")
	cmdFlags.Parse(osArgs)

	if appName == "" {
		cmd.result.Errorf("app is a mandatory argument")
	} else {
		cmd.appName = appName
	}

	if envName == "" {
		cmd.result.Errorf("env is a mandatory argument")
	} else {
		cmd.envName = envName
	}

	if keep < 2 {
		cmd.result.Errorf("keep must be at least 2")
	} else {
		cmd.keep = keep
	}

	return cmd.result
}

func (cmd *Set) Exec() *Result {

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

	// Retrieve the repository index.
	if ri, err := getRepoIndex(stg, cmd.appName); err == nil {

		// Retrieve and update the environment.
		if env, err := ri.GetEnv(cmd.envName); err != nil {
			cmd.result.AppendError(err)
			return cmd.result
		} else {
			env.SetKeep(cmd.keep)
			if err := ri.SetEnv(cmd.envName, env); err != nil {
				cmd.result.AppendError(err)
				return cmd.result
			}
		}

		// Write the index back.
		if err := setRepoIndex(stg, ri); err != nil {
			cmd.result.AppendError(err)
		}
	} else {
		cmd.result.AppendError(err)
	}

	return cmd.result
}

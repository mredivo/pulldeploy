package command

import (
	"flag"
	"fmt"

	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/storage"
)

// pulldeploy status -app=<app>
type Status struct {
	result  *Result
	pdcfg   pdconfig.PDConfig
	appName string
}

func (cmd *Status) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *Result {

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

	return cmd.result
}

func (cmd *Status) Exec() *Result {

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

		// Print a summary of the state of the application.
		// TODO: replace this simple JSON dump
		if text, err := ri.ToJSON(); err == nil {
			fmt.Println(string(text))
		} else {
			cmd.result.AppendError(err)
			return cmd.result
		}
	} else {
		cmd.result.AppendError(err)
	}

	return cmd.result
}

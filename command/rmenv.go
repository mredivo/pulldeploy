package command

import (
	"flag"

	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/storage"
)

// pulldeploy rmenv -app=<app> envname [envname envname ...]
type Rmenv struct {
	el       *ErrorList
	pdcfg    pdconfig.PDConfig
	appName  string
	envNames []string
}

func (cmd *Rmenv) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

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

	if len(cmdFlags.Args()) < 1 {
		cmd.el.Errorf("at least 1 environment name must be specified")
	} else {
		cmd.envNames = cmdFlags.Args()
	}

	return cmd.el
}

func (cmd *Rmenv) Exec() *ErrorList {

	// Ensure the app definition exists.
	if _, err := cmd.pdcfg.GetAppConfig(cmd.appName); err != nil {
		cmd.el.Append(err)
		return cmd.el
	}

	// Get access to the repo storage.
	stgcfg := cmd.pdcfg.GetStorageConfig()
	stg, err := storage.New(storage.StorageType(stgcfg.Type), stgcfg.Params)
	if err != nil {
		cmd.el.Append(err)
		return cmd.el
	}

	// Retrieve the repository index and update it.
	if ri, err := getRepoIndex(stg, cmd.appName); err == nil {

		successCount := 0
		for _, envName := range cmd.envNames {
			if err := ri.RmEnv(envName); err != nil {
				cmd.el.Append(err)
			} else {
				successCount++
			}
		}
		if successCount == 0 {
			return cmd.el
		}

		if err := setRepoIndex(stg, ri); err != nil {
			cmd.el.Append(err)
		}
	} else {
		cmd.el.Append(err)
	}

	return cmd.el
}

package command

import (
	"flag"

	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/storage"
)

// pulldeploy disable -app=<app> -version=<version>
type Disable struct {
	result     *Result
	pdcfg      pdconfig.PDConfig
	appName    string
	appVersion string
}

func (cmd *Disable) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *Result {

	var appName, appVersion string
	cmd.result = NewResult(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&appVersion, "version", "", "version of the application being disabled")
	cmdFlags.Parse(osArgs)

	if appName == "" {
		cmd.result.Errorf("app is a mandatory argument")
	} else {
		cmd.appName = appName
	}

	if appVersion == "" {
		cmd.result.Errorf("version is a mandatory argument")
	} else {
		cmd.appVersion = appVersion
	}

	return cmd.result
}

func (cmd *Disable) Exec() *Result {

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

		// Retrieve and update the version.
		if vers, err := ri.GetVersion(cmd.appVersion); err != nil {
			cmd.result.AppendError(err)
			return cmd.result
		} else {
			vers.Disable()
			if err := ri.SetVersion(cmd.appVersion, vers); err != nil {
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

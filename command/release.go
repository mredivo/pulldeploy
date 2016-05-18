package command

import (
	"flag"

	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/signaller"
	"github.com/mredivo/pulldeploy/storage"
)

// pulldeploy release -app=<app> -version=<version> -env=<env> [host1, host2, ...]
type Release struct {
	el         *ErrorList
	pdcfg      pdconfig.PDConfig
	appName    string
	appVersion string
	envName    string
	hosts      []string
}

func (cmd *Release) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

	var appName, appVersion, envName string
	cmd.el = NewErrorList(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&appVersion, "version", "", "version of the application to be released")
	cmdFlags.StringVar(&envName, "env", "", "environment in which to release")
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

	if envName == "" {
		cmd.el.Errorf("env is a mandatory argument")
	} else {
		cmd.envName = envName
	}

	cmd.hosts = cmdFlags.Args()

	return cmd.el
}

func (cmd *Release) Exec() *ErrorList {

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

	// Open the signaller, for notifying the pulldeploy daemons.
	sgnlr := signaller.New(cmd.pdcfg.GetSignallerConfig())
	sgnlr.Open()
	defer sgnlr.Close()

	// Retrieve the repository index.
	if ri, err := getRepoIndex(stg, cmd.appName); err == nil {

		// Retrieve the environment.
		if env, err := ri.GetEnv(cmd.envName); err != nil {
			cmd.el.Append(err)
			return cmd.el
		} else {

			// Indicate that this is the currently active version.
			if err := env.Release(cmd.appVersion, cmd.hosts); err != nil {
				cmd.el.Append(err)
				return cmd.el
			}

			// Put the updated environment back into the index.
			if err := ri.SetEnv(cmd.envName, env); err != nil {
				cmd.el.Append(err)
				return cmd.el
			}
		}

		// Write the index back.
		if err := setRepoIndex(stg, ri); err != nil {
			cmd.el.Append(err)
			return cmd.el
		}

		// Send out a notification.
		sgnlr.Notify(cmd.envName, cmd.appName, []byte{})

	} else {
		cmd.el.Append(err)
	}

	return cmd.el
}

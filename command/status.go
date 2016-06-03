package command

import (
	"flag"
	"fmt"
	"sort"
	"time"

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
	appCfg, err := cmd.pdcfg.GetAppConfig(cmd.appName)
	if err != nil {
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
		fmt.Printf("%s (%q) Status:\n", appCfg.Description, cmd.appName)

		// Order the environments alphabetically.
		var envs []string
		for envName, _ := range ri.Envs {
			envs = append(envs, envName)
		}
		sort.Strings(envs)

		// Iterate over the environments.
		for _, envName := range envs {
			v, _ := ri.GetEnv(envName)
			fmt.Printf("  %s:\n", envName)
			if v.Preview != "" {
				fmt.Printf("    Keep: %2d Current Version: %q Prior Version: %q Preview Version: %q\n",
					v.Keep, v.Current, v.Prior, v.Preview)
				fmt.Printf("      Preview Hosts:\n")
				for _, hostName := range v.Previewers {
					fmt.Printf("        %s\n", hostName)
				}
			} else {
				fmt.Printf("    Keep: %2d Current Version: %q Prior Version: %q\n",
					v.Keep, v.Current, v.Prior)
			}
			if len(v.Deployed) > 0 {
				fmt.Printf("    Deploy History:\n")
				for _, histEvent := range v.Deployed {
					fmt.Printf("      %s on %s\n", histEvent.Version, histEvent.TS.Format(time.RFC1123))
				}
			}
			if len(v.Released) > 0 {
				fmt.Printf("    Release History:\n")
				for _, histEvent := range v.Released {
					fmt.Printf("      %s on %s\n", histEvent.Version, histEvent.TS.Format(time.RFC1123))
				}
			}
		}

		// Iterate over the versions.
		fmt.Printf("  Uploaded Versions:\n")
		for _, v := range ri.VersionList() {
			released := "no "
			if v.Released {
				released = "yes"
			}
			disabled := ""
			if !v.Enabled {
				disabled = "  DISABLED"
			}
			fmt.Printf("      %s on %s  Released: %s%s\n", v.Name, v.TS.Format(time.RFC1123), released, disabled)
		}

	} else {
		cmd.result.AppendError(err)
	}

	return cmd.result
}

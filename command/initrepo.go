package command

import (
	"flag"
	"fmt"

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

	cmdFlags := flag.NewFlagSet("initrepo", flag.ExitOnError)
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

func (cmd *Initrepo) Exec() {
	placeHolder("initrepo(%s, %d)\n", cmd.appName, cmd.keep)

	// Ensure the app definition exists.
	if _, err := cmd.pdcfg.GetAppConfig(cmd.appName); err != nil {
		fmt.Printf("Repository initialization error: %s\n", err.Error())
		return
	}

	// Get access to the repo storage.
	stgcfg := cmd.pdcfg.GetStorageConfig()
	stg, err := storage.NewStorage(stgcfg.Type, stgcfg.Params)
	if err != nil {
		fmt.Printf("Repository initialization error: %s\n", err.Error())
		return
	}

	// Initialize the index and store it.
	ri := repo.NewRepoIndex(cmd.appName, cmd.keep)
	if text, err := ri.ToJSON(); err == nil {
		fmt.Println(string(text))
		if err := stg.Put(ri.RepoIndexPath(), text); err != nil {
			fmt.Printf("Repository initialization error: %s\n", err.Error())
		}
	}
}

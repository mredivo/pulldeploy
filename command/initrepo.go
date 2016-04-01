package command

import (
	"flag"
	"fmt"

	cfg "github.com/mredivo/pulldeploy/configloader"
	"github.com/mredivo/pulldeploy/repo"
	"github.com/mredivo/pulldeploy/storage"
)

// pulldeploy initrepo -app=<app> [-keep=n]
type Initrepo struct {
	appName string
	keep    int
}

func (cmd *Initrepo) CheckArgs(osArgs []string) bool {

	var appName string
	var keep int

	cmdFlags := flag.NewFlagSet("initrepo", flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application to create in the repository")
	cmdFlags.IntVar(&keep, "keep", 5, "the number of versions of app to keep in the repository")
	cmdFlags.Parse(osArgs)

	isValid := true

	if appName == "" {
		fmt.Println("app is a mandatory argument")
		isValid = false
	} else {
		cmd.appName = appName
	}

	if keep < 2 {
		fmt.Println("keep must be at least 2")
		isValid = false
	} else {
		cmd.keep = keep
	}

	return isValid
}

func (cmd *Initrepo) Exec() {
	fmt.Printf("initrepo(%s, %d)\n", cmd.appName, cmd.keep)

	stgcfg := cfg.GetStorageConfig()
	stg, err := storage.NewStorage(stgcfg.Type, stgcfg.Params)
	if err != nil {
		fmt.Printf("Repository initialization error: %s\n", err.Error())
		return
	}

	ri := repo.NewRepoIndex(cmd.appName, cmd.keep)
	if text, err := ri.ToJSON(); err == nil {
		fmt.Println(string(text))
		if err := stg.Put(ri.RepoIndexPath(), text); err != nil {
			fmt.Printf("Repository initialization error: %s\n", err.Error())
		}
	}
}

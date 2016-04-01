package command

import (
	"flag"
	"fmt"
	"sort"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// pulldeploy list
type List struct {
	el    *ErrorList
	pdcfg pdconfig.PDConfig
}

func (cmd *List) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

	cmd.el = NewErrorList(cmdName)
	cmd.pdcfg = pdcfg

	// Define a null set, so we can complain about extraneous args.
	cmdFlags := flag.NewFlagSet("list", flag.ExitOnError)
	cmdFlags.Parse(osArgs)

	return cmd.el
}

func (cmd *List) Exec() {

	// Fetch the list of applications.
	appList := cmd.pdcfg.GetAppList()

	// Extract the app names, and sort them alphabetically.
	var keys []string
	for k, _ := range appList {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Print each definition, in alphabetical order.
	for _, appName := range keys {

		fmt.Printf("%s\n", appName)

		appConfig := appList[appName]
		fmt.Printf("    Description : %s\n", appConfig.Description)
		fmt.Printf("    Secret      : %s\n", appConfig.Secret)
		fmt.Printf("    Directory   : %s\n", appConfig.Directory)
		fmt.Printf("    User        : %s\n", appConfig.User)
		fmt.Printf("    Group       : %s\n", appConfig.Group)
	}
}

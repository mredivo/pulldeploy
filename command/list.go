package command

import (
	"fmt"
	"sort"

	cfg "github.com/mredivo/pulldeploy/configloader"
)

// pulldeploy list
type List struct {
}

func (cmd *List) CheckArgs() bool {
	return true
}

func (cmd *List) Exec() {

	// Fetch the list of applications.
	appList := cfg.GetAppList()

	// Extract the app names, and sort them alphabetically.
	var keys []string
	for k, _ := range appList {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Print each definition, in alphabetical order.
	for _, appName := range keys {

		fmt.Printf("%s\n", appName)

		v := appList[appName]
		switch v.(type) {
		case *cfg.AppConfig:
			appConfig := v.(*cfg.AppConfig)
			fmt.Printf("    Description : %s\n", appConfig.Description)
			fmt.Printf("    Secret      : %s\n", appConfig.Secret)
			fmt.Printf("    Directory   : %s\n", appConfig.Directory)
			fmt.Printf("    User        : %s\n", appConfig.User)
			fmt.Printf("    Group       : %s\n", appConfig.Group)
		case error:
			err := v.(error)
			fmt.Printf("    Invalid application configuration: %s\n", err.Error())
		default:
			fmt.Printf("    Logic error: GetAppList() returned an unexpected type\n")
		}
	}
}

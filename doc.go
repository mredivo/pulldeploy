/*
Pulldeploy manages a repository of single-artifact application deployment packages, and
deploys and releases those packages on their target hosts.

usage: pulldeploy <command> [<args>]

Commands:

    Help:
        -h, help [<command>]
        -v, version

    Repository management:
        pulldeploy initrepo -app=<app>
        pulldeploy addenv   -app=<app> envname [envname envname ...]
        pulldeploy rmenv    -app=<app> envname [envname envname ...]
        pulldeploy set      -app=<app> -env=<env> [-keep=n]

    Release management:
        pulldeploy upload  -app=<app> -version=<version> [-disabled] <file>
        pulldeploy enable  -app=<app> -version=<version>
        pulldeploy disable -app=<app> -version=<version>
        pulldeploy purge   -app=<app> -version=<version>
        pulldeploy deploy  -app=<app> -version=<version> -env=<env>
        pulldeploy release -app=<app> -version=<version> -env=<env> [host1, host2, ...]

    Informational:
        pulldeploy list
        pulldeploy status -app=<app>
        pulldeploy listhosts -app=<app> -env=<env>

    Daemon:
        pulldeploy daemon -env=<env> [-logfile=<logfilename>]
*/
package main

import "fmt"

const usageShort = `usage: pulldeploy <command> [<args>]
    help - show list of commands`

const usageLong = `
usage: pulldeploy <command> [<args>]

Commands:

    Help:
        -h, help [<command>]
        -v, version

    Repository management:
        pulldeploy initrepo -app=<app>
        pulldeploy addenv   -app=<app> envname [envname envname ...]
        pulldeploy rmenv    -app=<app> envname [envname envname ...]
        pulldeploy set      -app=<app> -env=<env> [-keep=n]

    Release management:
        pulldeploy upload  -app=<app> -version=<version> [-disabled] <file>
        pulldeploy enable  -app=<app> -version=<version>
        pulldeploy disable -app=<app> -version=<version>
        pulldeploy purge   -app=<app> -version=<version>
        pulldeploy deploy  -app=<app> -version=<version> -env=<env>
        pulldeploy release -app=<app> -version=<version> -env=<env> [host1, host2, ...]

    Informational:
        pulldeploy list
        pulldeploy status -app=<app>
        pulldeploy listhosts -app=<app> -env=<env>

    Daemon:
        pulldeploy daemon -env=<env> [-logfile=<logfilename>]
`

func showCommandHelp(command string) bool {
	isValid := true
	switch command {
	case "initrepo":
		fmt.Println("usage: pulldeploy initrepo -app=<app>")
	case "addenv":
		fmt.Println("usage: pulldeploy addenv -app=<app> envname [envname envname ...]")
	case "rmenv":
		fmt.Println("usage: pulldeploy rmenv -app=<app> envname [envname envname ...]")
	case "set":
		fmt.Println("usage: pulldeploy set -app=<app> -env=<env> [-keep=n]")
	case "upload":
		fmt.Println("usage: pulldeploy upload -app=<app> -version=<version> [-disabled] <file>")
	case "enable":
		fmt.Println("usage: pulldeploy enable -app=<app> -version=<version>")
	case "disable":
		fmt.Println("usage: pulldeploy disable -app=<app> -version=<version>")
	case "purge":
		fmt.Println("usage: pulldeploy purge -app=<app> -version=<version>")
	case "deploy":
		fmt.Println("usage: pulldeploy deploy -app=<app> -version=<version> -env=<env>")
	case "release":
		fmt.Println("usage: pulldeploy release -app=<app> -version=<version> -env=<env> [host1, host2, ...]")
	case "list":
		fmt.Println("usage: pulldeploy list")
	case "status":
		fmt.Println("usage: pulldeploy status -app=<app>")
	case "listhosts":
		fmt.Println("usage: pulldeploy listhosts -app=<app> -env=<env>")
	case "daemon":
		fmt.Println("usage: pulldeploy daemon -env=<env> [-logfile=<logfilename>]")
	default:
		fmt.Printf("invalid command: %q\n", command)
		isValid = false
	}
	return isValid
}

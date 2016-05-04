/** PullDeploy -- A Release Manager for Single-Artifact Deploys.
 */
package main

import (
	"fmt"
	"os"

	"github.com/mredivo/pulldeploy/command"
	"github.com/mredivo/pulldeploy/pdconfig"
)

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
        pulldeploy daemon -env=<env> <daemon args>...
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
		fmt.Println("usage: pulldeploy daemon -env=<env> <daemon args>...")
	default:
		fmt.Printf("invalid command: %q\n", command)
		isValid = false
	}
	return isValid
}

func main() {

	// Ensure there are at least two command-line arguments.
	if len(os.Args) < 2 {
		fmt.Println(usageShort)
		os.Exit(1)
	}

	// Load the configuration.
	var pdcfg pdconfig.PDConfig
	var errs []error
	if pdcfg, errs = pdconfig.LoadPulldeployConfig(); pdcfg == nil {
		// Not loading a configuration in pdcfg is fatal.
		for _, err := range errs {
			fmt.Println(err.Error())
		}
		os.Exit(3)
	}
	// If we got a configuration, remaining errors are warnings.
	for _, err := range errs {
		fmt.Println(err.Error())
	}

	// Parse the command line appropriately for the given subcommand.
	var cmd command.Handler
	switch os.Args[1] {
	case "help", "-h", "-help", "--help":
		if len(os.Args) > 2 {
			if isValidCommand := showCommandHelp(os.Args[2]); !isValidCommand {
				fmt.Println(usageLong)
			}
		} else {
			fmt.Println(usageLong)
		}
	case "version", "-v", "-version", "--version":
		fmt.Println(versionInfo.OneLine())
	case "initrepo":
		cmd = new(command.Initrepo)
	case "set":
		cmd = new(command.Set)
	case "addenv":
		cmd = new(command.Addenv)
	case "rmenv":
		cmd = new(command.Rmenv)
	case "upload":
		cmd = new(command.Upload)
	case "enable":
		cmd = new(command.Enable)
	case "disable":
		cmd = new(command.Disable)
	case "purge":
		cmd = new(command.Purge)
	case "deploy":
		cmd = new(command.Deploy)
	case "release":
		cmd = new(command.Release)
	case "list":
		cmd = new(command.List)
	case "status":
		cmd = new(command.Status)
	case "listhosts":
		cmd = new(command.Listhosts)
	case "daemon":
		cmd = new(command.Daemon)
	default:
		fmt.Printf("%q is not a valid command\n", os.Args[1])
		os.Exit(2)
	}

	// If a command was recognized, validate and execute it.
	var exitCode int
	if cmd != nil {
		if el := cmd.CheckArgs(os.Args[1], pdcfg, os.Args[2:]); el.Len() == 0 {
			el = cmd.Exec()
			for _, s := range el.Errors() {
				fmt.Println(s)
				exitCode = 4
			}
		} else {
			exitCode = 2
			for _, s := range el.Errors() {
				fmt.Println(s)
			}
			showCommandHelp(os.Args[1])
		}
	}

	if exitCode > 0 {
		os.Exit(exitCode)
	}
}

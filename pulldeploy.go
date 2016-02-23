/** PullDeploy -- A Release Manager for Single-Artifact Deploys.
 */
package main

import (
	"flag"
	"fmt"
	"os"
)

const usageShort = `usage: pulldeploy <command> [<args>]
    help - show list of commands`

const usageLong = `
usage: pulldeploy <command> [<args>]

Commands:

    Help:
        help [<command>]

    Repository management:
        pulldeploy initrepo -app=<app> [-keep=n]
        pulldeploy set      -app=<app> [-keep=n]
        pulldeploy addenv   -app=<app> envname [envname envname ...]
        pulldeploy rmenv    -app=<app> envname [envname envname ...]

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

func main() {

	// Variables sourced from the command line.
	var appName string    // The name of the application on which to act
	var appVersion string // The application version
	var envName string    // The name of the environment in which to act
	var keep int          // The number of versions of an app to keep in the repo

	// Ensure there are at least two command-line arguments.
	if len(os.Args) < 2 {
		fmt.Println(usageShort)
		return
	}

	// Parse the command line appropriately for the given subcommand.
	switch os.Args[1] {

	case "help", "-h":
		if len(os.Args) == 3 {
			if isValidCommand := showCommandHelp(os.Args[2]); !isValidCommand {
				fmt.Println(usageLong)
			}
		} else {
			fmt.Println(usageLong)
		}

	case "initrepo":
		cmdFlags := flag.NewFlagSet("initrepo", flag.ExitOnError)
		cmdFlags.StringVar(&appName, "app", "", "name of the application to create in the repository")
		cmdFlags.IntVar(&keep, "keep", 5, "the number of versions of app to keep in the repository")
		cmdFlags.Parse(os.Args[2:])
		cmd := cmdInitrepo{}
		if isValid := cmd.checkArgs(appName, keep); isValid {
			cmd.exec()
		} else {
			showCommandHelp(os.Args[1])
		}

	case "set":
		cmdFlags := flag.NewFlagSet("set", flag.ExitOnError)
		cmdFlags.StringVar(&appName, "app", "", "name of the application whose repository to update")
		cmdFlags.IntVar(&keep, "keep", 5, "the number of versions of app to keep in the repository")
		cmdFlags.Parse(os.Args[2:])
		cmd := cmdSet{}
		if isValid := cmd.checkArgs(appName, keep); isValid {
			cmd.exec()
		} else {
			showCommandHelp(os.Args[1])
		}

	case "addenv":
		cmdFlags := flag.NewFlagSet("addenv", flag.ExitOnError)
		cmdFlags.StringVar(&appName, "app", "", "name of the application")
		cmdFlags.Parse(os.Args[2:])
		cmd := cmdAddenv{}
		if isValid := cmd.checkArgs(appName, cmdFlags.Args()); isValid {
			cmd.exec()
		} else {
			showCommandHelp(os.Args[1])
		}

	case "rmenv":
		cmdFlags := flag.NewFlagSet("rmenv", flag.ExitOnError)
		cmdFlags.StringVar(&appName, "app", "", "name of the application")
		cmdFlags.Parse(os.Args[2:])
		cmd := cmdRmenv{}
		if isValid := cmd.checkArgs(appName, cmdFlags.Args()); isValid {
			cmd.exec()
		} else {
			showCommandHelp(os.Args[1])
		}

	case "upload":
		var disabled bool
		cmdFlags := flag.NewFlagSet("upload", flag.ExitOnError)
		cmdFlags.StringVar(&appName, "app", "", "name of the application")
		cmdFlags.StringVar(&appVersion, "version", "", "version of the application being uploaded")
		cmdFlags.BoolVar(&disabled, "disabled", false, "upload in disabled state")
		cmdFlags.Parse(os.Args[2:])
		cmd := cmdUpload{}
		if isValid := cmd.checkArgs(appName, appVersion, disabled, cmdFlags.Args()); isValid {
			cmd.exec()
		} else {
			showCommandHelp(os.Args[1])
		}

	case "enable":
		cmdFlags := flag.NewFlagSet("enable", flag.ExitOnError)
		cmdFlags.StringVar(&appName, "app", "", "name of the application")
		cmdFlags.StringVar(&appVersion, "version", "", "version of the application being enabled")
		cmdFlags.Parse(os.Args[2:])
		cmd := cmdEnable{}
		if isValid := cmd.checkArgs(appName, appVersion); isValid {
			cmd.exec()
		} else {
			showCommandHelp(os.Args[1])
		}

	case "disable":
		cmdFlags := flag.NewFlagSet("disable", flag.ExitOnError)
		cmdFlags.StringVar(&appName, "app", "", "name of the application")
		cmdFlags.StringVar(&appVersion, "version", "", "version of the application being disabled")
		cmdFlags.Parse(os.Args[2:])
		cmd := cmdDisable{}
		if isValid := cmd.checkArgs(appName, appVersion); isValid {
			cmd.exec()
		} else {
			showCommandHelp(os.Args[1])
		}

	case "purge":
		cmdFlags := flag.NewFlagSet("purge", flag.ExitOnError)
		cmdFlags.StringVar(&appName, "app", "", "name of the application")
		cmdFlags.StringVar(&appVersion, "version", "", "version of the application being purged")
		cmdFlags.Parse(os.Args[2:])
		cmd := cmdPurge{}
		if isValid := cmd.checkArgs(appName, appVersion); isValid {
			cmd.exec()
		} else {
			showCommandHelp(os.Args[1])
		}

	case "deploy":
		cmdFlags := flag.NewFlagSet("deploy", flag.ExitOnError)
		cmdFlags.StringVar(&appName, "app", "", "name of the application")
		cmdFlags.StringVar(&appVersion, "version", "", "version of the application to be deployed")
		cmdFlags.StringVar(&envName, "env", "", "environment to which to deploy")
		cmdFlags.Parse(os.Args[2:])
		cmd := cmdDeploy{}
		if isValid := cmd.checkArgs(appName, appVersion, envName); isValid {
			cmd.exec()
		} else {
			showCommandHelp(os.Args[1])
		}

	case "release":
		cmdFlags := flag.NewFlagSet("release", flag.ExitOnError)
		cmdFlags.StringVar(&appName, "app", "", "name of the application")
		cmdFlags.StringVar(&appVersion, "version", "", "version of the application to be released")
		cmdFlags.StringVar(&envName, "env", "", "environment in which to release")
		cmdFlags.Parse(os.Args[2:])
		cmd := cmdRelease{}
		if isValid := cmd.checkArgs(appName, appVersion, envName, cmdFlags.Args()); isValid {
			cmd.exec()
		} else {
			showCommandHelp(os.Args[1])
		}

	case "list":
		cmdFlags := flag.NewFlagSet("list", flag.ExitOnError)
		cmdFlags.Parse(os.Args[2:])
		cmd := cmdList{}
		if isValid := cmd.checkArgs(); isValid {
			cmd.exec()
		} else {
			showCommandHelp(os.Args[1])
		}

	case "status":
		cmdFlags := flag.NewFlagSet("status", flag.ExitOnError)
		cmdFlags.StringVar(&appName, "app", "", "name of the application")
		cmdFlags.Parse(os.Args[2:])
		cmd := cmdStatus{}
		if isValid := cmd.checkArgs(appName); isValid {
			cmd.exec()
		} else {
			showCommandHelp(os.Args[1])
		}

	case "listhosts":
		cmdFlags := flag.NewFlagSet("listhosts", flag.ExitOnError)
		cmdFlags.StringVar(&appName, "app", "", "name of the application")
		cmdFlags.StringVar(&envName, "env", "", "environment in which to release")
		cmdFlags.Parse(os.Args[2:])
		cmd := cmdListhosts{}
		if isValid := cmd.checkArgs(appName, envName); isValid {
			cmd.exec()
		} else {
			showCommandHelp(os.Args[1])
		}

	case "daemon":
		// Enter the monitoring loop and put changes into effect locally.
		cmdFlags := flag.NewFlagSet("daemon", flag.ExitOnError)
		cmdFlags.StringVar(&envName, "env", "", "environment to be monitored")
		cmdFlags.Parse(os.Args[2:])

	default:
		fmt.Printf("%q is not a valid command\n", os.Args[1])
		os.Exit(2)
	}

}

func showCommandHelp(command string) bool {
	isValid := true
	switch command {
	case "initrepo":
		fmt.Println("usage: pulldeploy initrepo -app=<app> [-keep=n]")
	case "set":
		fmt.Println("usage: pulldeploy set -app=<app> [-keep=n]")
	case "addenv":
		fmt.Println("usage: pulldeploy addenv -app=<app> envname [envname envname ...]")
	case "rmenv":
		fmt.Println("usage: pulldeploy rmenv -app=<app> envname [envname envname ...]")
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

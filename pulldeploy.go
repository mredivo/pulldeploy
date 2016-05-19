package main

import (
	"fmt"
	"os"

	"github.com/mredivo/pulldeploy/command"
	"github.com/mredivo/pulldeploy/pdconfig"
)

func main() {

	// Ensure there are at least two command-line arguments.
	if len(os.Args) < 2 {
		fmt.Println(usageShort)
		os.Exit(1)
	}

	// Load the configuration.
	var pdcfg pdconfig.PDConfig
	var errs []error
	if pdcfg, errs = pdconfig.LoadPulldeployConfig(""); pdcfg == nil {
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
		fmt.Println(pdcfg.GetVersionInfo().OneLine())
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

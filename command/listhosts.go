package command

import (
	"flag"
	"fmt"

	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/signaller"
)

// pulldeploy listhosts -app=<app> -env=<env>
type Listhosts struct {
	el      *ErrorList
	pdcfg   pdconfig.PDConfig
	appName string
	envName string
}

func (cmd *Listhosts) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

	var appName, envName string
	cmd.el = NewErrorList(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&envName, "env", "", "environment in which to release")
	cmdFlags.Parse(osArgs)

	if appName == "" {
		cmd.el.Errorf("app is a mandatory argument")
	} else {
		cmd.appName = appName
	}

	if envName == "" {
		cmd.el.Errorf("env is a mandatory argument")
	} else {
		cmd.envName = envName
	}

	return cmd.el
}

func (cmd *Listhosts) Exec() *ErrorList {

	// Ensure the app definition exists.
	if _, err := cmd.pdcfg.GetAppConfig(cmd.appName); err != nil {
		cmd.el.Append(err)
		return cmd.el
	}

	// Open the signaller, for access to the hosts registry.
	sgnlr := signaller.New(cmd.pdcfg.GetSignallerConfig())
	sgnlr.Open()
	defer sgnlr.Close()

	// TODO: Offer a machine-readable format.
	fmt.Printf("Registered %q hosts in %q\n", cmd.appName, cmd.envName)
	hr := sgnlr.GetRegistry()
	var count int
	for _, v := range hr.Hosts(cmd.envName, cmd.appName) {
		fmt.Printf("   Host: %q Version: %q\n", v.Hostname, v.AppVersion)
		count++
	}
	if count == 1 {
		fmt.Printf("%d host\n", count)
	} else {
		fmt.Printf("%d hosts\n", count)
	}

	return cmd.el
}

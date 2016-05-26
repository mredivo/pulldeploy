package command

import (
	"flag"
	"fmt"

	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/signaller"
)

// pulldeploy listhosts -app=<app> -env=<env>
type Listhosts struct {
	result  *Result
	pdcfg   pdconfig.PDConfig
	appName string
	envName string
}

func (cmd *Listhosts) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *Result {

	var appName, envName string
	cmd.result = NewResult(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&envName, "env", "", "environment in which to release")
	cmdFlags.Parse(osArgs)

	if appName == "" {
		cmd.result.Errorf("app is a mandatory argument")
	} else {
		cmd.appName = appName
	}

	if envName == "" {
		cmd.result.Errorf("env is a mandatory argument")
	} else {
		cmd.envName = envName
	}

	return cmd.result
}

func (cmd *Listhosts) Exec() *Result {

	// Ensure the app definition exists.
	if _, err := cmd.pdcfg.GetAppConfig(cmd.appName); err != nil {
		cmd.result.AppendError(err)
		return cmd.result
	}

	// Open the signaller, for access to the hosts registry.
	sgnlr := signaller.New(cmd.pdcfg.GetSignallerConfig(), nil)
	sgnlr.Open()
	defer sgnlr.Close()

	// Print the list.
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

	return cmd.result
}

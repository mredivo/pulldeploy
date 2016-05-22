package command

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/mredivo/pulldeploy/deployment"
	"github.com/mredivo/pulldeploy/logging"
	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/signaller"
	"github.com/mredivo/pulldeploy/storage"
)

// pulldeploy daemon ...
type Daemon struct {
	el         *ErrorList
	pdcfg      pdconfig.PDConfig
	envName    string
	lw         *logging.Writer
	stg        storage.Storage
	hr         *signaller.Registry
	myHostname string
}

func (cmd *Daemon) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

	var envName string
	cmd.el = NewErrorList(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmdFlags.StringVar(&envName, "env", "", "environment to be monitored")
	cmdFlags.Parse(osArgs)

	if envName == "" {
		cmd.el.Errorf("env is a mandatory argument")
	} else {
		cmd.envName = envName
	}

	return cmd.el
}

func (cmd *Daemon) Exec() *ErrorList {

	// Set up reload and termination signal handlers.
	var sigterm = make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	var sighup = make(chan os.Signal, 1)
	signal.Notify(sighup, syscall.SIGHUP)
	var sigusr1 = make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	// Open the logger.
	logger := logging.New("pulldeploy", "", true)
	defer logger.Close()
	cmd.lw = logger.GetWriter("", "debug")
	cmd.lw.Info(cmd.pdcfg.GetVersionInfo().OneLine())

	// Instantiate the signaller that tells us when apps need attention.
	sgnlr := signaller.New(cmd.pdcfg.GetSignallerConfig())
	appEvent := sgnlr.Open()
	defer sgnlr.Close()
	cmd.hr = sgnlr.GetRegistry()

	// Get access to the repo storage.
	stgcfg := cmd.pdcfg.GetStorageConfig()
	if stg, err := storage.New(storage.AccessMethod(stgcfg.AccessMethod), stgcfg.Params); err == nil {
		cmd.stg = stg
	} else {
		cmd.el.Append(err)
		return cmd.el
	}

	// Determine the local hostname.
	cmd.myHostname, _ = os.Hostname()
	cmd.lw.Info("Registering host name %q", cmd.myHostname)

	// Get the set of applications to monitor.
	appList := cmd.pdcfg.GetAppList()

	var registerAppHosts = func() {
		for appName, _ := range appList {

			// Retrieve the app definition.
			appCfg, err := cmd.pdcfg.GetAppConfig(appName)
			if err != nil {
				cmd.lw.Error("Error getting configuration for %q: %s", appName, err.Error())
				continue
			}

			// Instantiate the deployment object for this application.
			dplmt, err := deployment.New(appName, appCfg)
			if err != nil {
				cmd.lw.Error("Error in deployment for %q: %s", appName, err.Error())
				continue
			}

			// Register with current version, and ask for notifications.
			cmd.hr.Register(cmd.envName, appName, cmd.myHostname, dplmt.GetCurrentLink())
			sgnlr.Monitor(cmd.envName, appName)
		}
	}

	var unregisterAppHosts = func() {
		for appName, _ := range appList {
			cmd.hr.Unregister(cmd.envName, appName, cmd.myHostname)
		}
	}

	var synchronize = func() {
		for appName, _ := range appList {
			cmd.synchronize(signaller.Notification{Source: signaller.KNS_FORCED, Appname: appName})
		}
	}

	// Register the host for each application in this environment.
	registerAppHosts()
	cmd.lw.Info("Startup complete")

	// Perform the initial synchronization with the repo.
	synchronize()

	// Processing loop.
	repeat := true
	for repeat {
		select {

		case appNotification := <-appEvent:
			// Make the local deploy/release state of the app match the repo index.
			cmd.synchronize(appNotification)

		case <-sigusr1:
			// Close and re-open the logfile.
			cmd.lw.Info("Received SIGUSR1")
			logger.OnRotate()

		case <-sighup:
			// Refresh the set of applications to monitor.
			cmd.lw.Info("Received SIGHUP")
			// Unregister and stop monitoring.
			unregisterAppHosts()
			cmd.pdcfg.RefreshAppList()
			appList = cmd.pdcfg.GetAppList()
			// Re-register and restart monitoring.
			registerAppHosts()
			synchronize()

		case <-sigterm:
			// Gracefully shut down.
			cmd.lw.Info("Received SIGINT/SIGTERM")
			repeat = false // Break loop.
		}
	}

	unregisterAppHosts()
	cmd.lw.Info("Termination complete")

	return cmd.el
}

func (cmd *Daemon) synchronize(an signaller.Notification) {

	cmd.lw.Info("Synchronizing %q in %q (%s)", an.Appname, cmd.envName, an.Source)

	// Retrieve the app definition.
	appCfg, err := cmd.pdcfg.GetAppConfig(an.Appname)
	if err != nil {
		cmd.lw.Error("Error getting configuration for %q: %s", an.Appname, err.Error())
		return
	}

	// Instantiate the deployment object for this application.
	dplmt, err := deployment.New(an.Appname, appCfg)
	if err != nil {
		cmd.lw.Error("Error in deployment for %q: %s", an.Appname, err.Error())
		return
	}

	// Retrieve the repository index.
	if ri, err := getRepoIndex(cmd.stg, an.Appname); err == nil {

		// Retrieve the environment.
		if env, err := ri.GetEnv(cmd.envName); err != nil {
			cmd.lw.Error("Error getting %q environment for %q: %s", cmd.envName, an.Appname, err.Error())
			return
		} else {

			// Determine whether any new versions have been deployed since we last checked.
			localVersionList := dplmt.ListVersions()
			newDeployments := subtractArray(env.Deployed, localVersionList)
			cmd.lw.Debug("Deployments for %s in %s: local=%v, repo=%v new=%v",
				an.Appname, cmd.envName, localVersionList, env.Deployed, newDeployments)

			// Fetch and unpack all new deployments.
			for _, version := range newDeployments {

				// Determine the base filename.
				filename := ri.ArtifactFilename(version, appCfg.ArtifactType)

				// Retrieve the artifact for that filename.
				if !dplmt.ArtifactPresent(version) {
					if art, err := cmd.stg.GetReader(ri.ArtifactPath(filename)); err == nil {
						if err := dplmt.WriteArtifact(version, art); err == nil {
							cmd.lw.Debug("Fetched artifact %q for %s in %s",
								ri.ArtifactPath(filename), cmd.envName, an.Appname)
						} else {
							cmd.lw.Error("Error writing artifact %q for %s in %s: %s",
								ri.ArtifactPath(filename), cmd.envName, an.Appname, err.Error())
							continue
						}
					} else {
						cmd.lw.Error("Error getting artifact %q for %s in %s: %s",
							ri.ArtifactPath(filename), cmd.envName, an.Appname, err.Error())
						continue
					}
				}

				// Retrieve the HMAC for that filename.
				if !dplmt.HMACPresent(version) {
					if hmac, err := cmd.stg.Get(ri.HMACPath(filename)); err == nil {
						if err := dplmt.WriteHMAC(version, hmac); err == nil {
							cmd.lw.Debug("Fetched HMAC %q for %s in %s",
								ri.HMACPath(filename), cmd.envName, an.Appname)
						} else {
							cmd.lw.Error("Error writing HMAC %q for %s in %s: %s",
								ri.HMACPath(filename), cmd.envName, an.Appname, err.Error())
							continue
						}
					} else {
						cmd.lw.Error("Error getting HMAC %q for %s in %s: %s",
							ri.HMACPath(filename), cmd.envName, an.Appname, err.Error())
						continue
					}
				}

				// Compare the calculated HMAC with the retrieved HMAC.
				if err := dplmt.CheckHMAC(version); err != nil {
					cmd.lw.Error("HMAC comparison FAILED for %s in %s, version %q",
						an.Appname, cmd.envName, version)
					continue
				}

				// Extract the artifact to the release directory.
				if err := dplmt.Extract(version); err != nil {
					cmd.lw.Error("Extract FAILED for %s in %s, version %q",
						an.Appname, cmd.envName, version)
					continue
				}

				// Execute the post-deploy command.
				cmd.logPostCommand(dplmt.PostDeploy(version))
			}

			// Determine the currently released version on the local host, and
			// update if necessary.
			localRelease := dplmt.GetCurrentLink()
			currentRelease := env.GetCurrentVersion(cmd.myHostname)
			cmd.lw.Debug("Current release: local=%q, repo=%q", localRelease, currentRelease)
			if localRelease != currentRelease && currentRelease != "" {
				if err := dplmt.Link(currentRelease); err == nil {
					cmd.lw.Info("Current release for %s in %s set to %q",
						an.Appname, cmd.envName, currentRelease)
					// Execute the post-release command.
					cmd.logPostCommand(dplmt.PostRelease(currentRelease))
					cmd.hr.Unregister(cmd.envName, an.Appname, cmd.myHostname)
					cmd.hr.Register(cmd.envName, an.Appname, cmd.myHostname, dplmt.GetCurrentLink())
				} else {
					cmd.lw.Error("Error setting current release for %s in %s to %q: %s",
						an.Appname, cmd.envName, currentRelease, err.Error())
				}
			}
		}

	} else {
		cmd.lw.Error("Error getting repo index for %q: %s", an.Appname, err.Error())
	}
}

func (cmd *Daemon) logPostCommand(command, curdir, stdout, stderr string) {
	if command != "" {
		if stdout == "" {
			if stderr == "" {
				cmd.lw.Info("Executed %q in %s", command, curdir)
			} else {
				cmd.lw.Info("Executed %q in %s\nstderr=%q", command, curdir, stderr)
				cmd.lw.Warn(stderr)
			}
		} else {
			if stderr == "" {
				cmd.lw.Info("Executed %q in %s\nstdout=%q", command, curdir, stdout)
			} else {
				cmd.lw.Info("Executed %q in %s\nstdout=%q\nstderr=%q", command, curdir, stdout, stderr)
				cmd.lw.Warn(stderr)
			}
		}
	}
}

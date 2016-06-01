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
	result     *Result
	pdcfg      pdconfig.PDConfig
	envName    string
	logFile    string
	lw         *logging.Writer
	stg        storage.Storage
	hr         *signaller.Registry
	myHostname string
	canary     map[string]int
}

func (cmd *Daemon) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *Result {

	var envName, logFile string
	cmd.result = NewResult(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmdFlags.StringVar(&envName, "env", "", "environment to be monitored")
	cmdFlags.StringVar(&logFile, "logfile", "", "name of log file (default stdout)")
	cmdFlags.Parse(osArgs)

	if envName == "" {
		cmd.result.Errorf("env is a mandatory argument")
	} else {
		cmd.envName = envName
	}
	cmd.logFile = logFile
	cmd.canary = make(map[string]int)

	return cmd.result
}

func (cmd *Daemon) Exec() *Result {

	// Set up reload and termination signal handlers.
	var sigterm = make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	var sighup = make(chan os.Signal, 1)
	signal.Notify(sighup, syscall.SIGHUP)
	var sigusr1 = make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	// Open the logger.
	logger := logging.New("pulldeploy", cmd.logFile, true)
	defer logger.Close()
	cmd.lw = logger.GetWriter("", cmd.pdcfg.GetLogLevel())
	cmd.lw.Info(cmd.pdcfg.GetVersionInfo().OneLine())

	// Instantiate the signaller that tells us when apps need attention.
	sgnlr := signaller.New(cmd.pdcfg.GetSignallerConfig(), cmd.lw)
	appEvent := sgnlr.Open()
	defer sgnlr.Close()
	cmd.hr = sgnlr.GetRegistry()

	// Get access to the repo storage.
	stgcfg := cmd.pdcfg.GetStorageConfig()
	if stg, err := storage.New(storage.AccessMethod(stgcfg.AccessMethod), stgcfg.Params); err == nil {
		cmd.stg = stg
	} else {
		cmd.result.AppendError(err)
		return cmd.result
	}

	// Determine the local hostname.
	cmd.myHostname, _ = os.Hostname()
	cmd.lw.Info("Host name: %q", cmd.myHostname)

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
			dplmt, err := deployment.New(appName, cmd.pdcfg, appCfg)
			if err != nil {
				cmd.lw.Error("Error in deployment for %q: %s", appName, err.Error())
				continue
			}

			// Register with current version, and ask for notifications.
			cmd.hr.Register(cmd.envName, appName, cmd.myHostname,
				dplmt.GetCurrentLink(), dplmt.GetDeployedVersions())
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

	return cmd.result
}

func (cmd *Daemon) synchronize(an signaller.Notification) {

	cmd.lw.Info("Synchronizing %q in %q (%s)", an.Appname, cmd.envName, an.Source)

	// Retrieve the app definition.
	appCfg, err := cmd.pdcfg.GetAppConfig(an.Appname)
	if err != nil {
		cmd.lw.Error("Error getting configuration for %q: %s", an.Appname, err.Error())
		return
	}

	// Get the extension for the artifact type.
	var extension string
	if ac, err := cmd.pdcfg.GetArtifactConfig(appCfg.ArtifactType); err == nil {
		extension = ac.Extension
	} else {
		cmd.lw.Error("Invalid ArtifactType for %q: %q", an.Appname, appCfg.ArtifactType)
		return
	}

	// Instantiate the deployment object for this application.
	dplmt, err := deployment.New(an.Appname, cmd.pdcfg, appCfg)
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

			// First check whether the repository has changed since we last looked.
			if canary, found := cmd.canary[an.Appname]; found && canary == ri.Canary {
				cmd.lw.Debug("Canary unchanged for %q: %d", an.Appname, canary)
				// Returning here misses the case where the repository has not changed,
				// but the local filesystem has. Log the message, but keep going.
				//return
			}

			// Determine whether any new versions have been deployed since we last checked.
			localVersionList := dplmt.GetDeployedVersions()
			var deployedVersionList []string
			for _, v := range env.Deployed {
				deployedVersionList = append(deployedVersionList, v.Version)
			}
			newDeployments := subtractArray(deployedVersionList, localVersionList)
			cmd.lw.Debug("Deployments for %s in %s: local=%v, repo=%v new=%v",
				an.Appname, cmd.envName, localVersionList, deployedVersionList, newDeployments)

			// Fetch and unpack all new deployments.
			for _, version := range newDeployments {

				// Determine the base filename.
				filename := ri.ArtifactFilename(version, extension)

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
				if err := dplmt.CheckHMAC(version); err == nil {
					cmd.lw.Debug("HMAC comparison succeeded for %s in %s, version %q",
						an.Appname, cmd.envName, version)
				} else {
					cmd.lw.Error("HMAC comparison FAILED for %s in %s, version %q",
						an.Appname, cmd.envName, version)
					continue
				}

				// Extract the artifact to the release directory.
				if err := dplmt.Extract(version); err == nil {
					cmd.lw.Debug("Extracted version %q for %s in %s",
						version, cmd.envName, an.Appname)
				} else {
					cmd.lw.Error("Extract FAILED for %s in %s, version %q: %s",
						an.Appname, cmd.envName, version, err.Error())
					continue
				}

				// Execute the post-deploy command.
				cmd.logPostCommand(dplmt.PostDeploy(version))
				cmd.hr.Register(cmd.envName, an.Appname, cmd.myHostname,
					dplmt.GetCurrentLink(), dplmt.GetDeployedVersions())
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
					cmd.hr.Register(cmd.envName, an.Appname, cmd.myHostname,
						dplmt.GetCurrentLink(), dplmt.GetDeployedVersions())
				} else {
					cmd.lw.Error("Error setting current release for %s in %s to %q: %s",
						an.Appname, cmd.envName, currentRelease, err.Error())
				}
			}

			// Note that the local host is in sync with the index.
			cmd.canary[an.Appname] = ri.Canary
		}

	} else {
		cmd.lw.Error("Error getting repo index for %q: %s", an.Appname, err.Error())
	}
}

func (cmd *Daemon) logPostCommand(cmdline string, err error) {
	if cmdline != "" {
		cmd.lw.Info(cmdline)
	}
	if err != nil {
		cmd.lw.Warn(err.Error())
	}
}

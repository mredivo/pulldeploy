package command

import (
	"flag"
	"os"

	"github.com/mredivo/pulldeploy/deployment"
	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/storage"
)

// pulldeploy upload -app=<app> -version=<version> [-disabled] <file>
type Upload struct {
	result     *Result
	pdcfg      pdconfig.PDConfig
	appName    string
	appVersion string
	disabled   bool
	filename   string
}

func (cmd *Upload) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *Result {

	var appName, appVersion string
	var disabled bool
	cmd.result = NewResult(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&appVersion, "version", "", "version of the application being uploaded")
	cmdFlags.BoolVar(&disabled, "disabled", false, "upload in disabled state")
	cmdFlags.Parse(osArgs)

	if appName == "" {
		cmd.result.Errorf("app is a mandatory argument")
	} else {
		cmd.appName = appName
	}

	if appVersion == "" {
		cmd.result.Errorf("version is a mandatory argument")
	} else {
		cmd.appVersion = appVersion
	}

	cmd.disabled = disabled

	if len(cmdFlags.Args()) < 1 {
		cmd.result.Errorf("filename is a mandatory argument")
	} else if len(cmdFlags.Args()) > 1 {
		cmd.result.Errorf("only one filename may be specified")
	} else {
		cmd.filename = cmdFlags.Args()[0]
	}

	return cmd.result
}

func (cmd *Upload) Exec() *Result {

	// Ensure the app definition exists.
	appCfg, err := cmd.pdcfg.GetAppConfig(cmd.appName)
	if err != nil {
		cmd.result.AppendError(err)
		return cmd.result
	}

	// Get access to the repo storage.
	stgcfg := cmd.pdcfg.GetStorageConfig()
	stg, err := storage.New(storage.AccessMethod(stgcfg.AccessMethod), stgcfg.Params)
	if err != nil {
		cmd.result.AppendError(err)
		return cmd.result
	}

	// Get the extension for the artifact type.
	var extension string
	if ac, err := cmd.pdcfg.GetArtifactConfig(appCfg.ArtifactType); err == nil {
		extension = ac.Extension
	} else {
		cmd.result.Errorf("Invalid ArtifactType for app: %q", appCfg.ArtifactType)
		return cmd.result
	}

	// Retrieve the repository index.
	if ri, err := getRepoIndex(stg, cmd.appName); err == nil {

		// Open the artifact to be uploaded.
		if fh, err := os.Open(cmd.filename); err == nil {
			defer fh.Close()

			// Determine the content length.
			fi, err := fh.Stat()
			if err != nil {
				cmd.result.AppendError(err)
				return cmd.result
			}

			// Write the artifact to the repo.
			repoFilename := ri.ArtifactFilename(cmd.appVersion, extension)
			repoPath := ri.ArtifactPath(repoFilename)
			if err := stg.PutReader(repoPath, fh, fi.Size()); err != nil {
				cmd.result.AppendError(err)
				return cmd.result
			}

			// Calculate the artifact HMAC and write that to the repo.
			if fh, err := os.Open(cmd.filename); err == nil {
				hmac := deployment.CalculateHMAC(fh, deployment.NewHMACCalculator(appCfg.Secret))
				hmacPath := ri.HMACPath(repoFilename)
				if err := stg.Put(hmacPath, hmac); err != nil {
					cmd.result.AppendError(err)
					return cmd.result
				}
			} else {
				cmd.result.AppendError(err)
				return cmd.result
			}

			// This callback will be called for each entry purged from the repository.
			onDelete := func(versionName string) {
				if vers, err := ri.GetVersion(versionName); err == nil {
					stg.Delete(ri.ArtifactPath(vers.Filename))
					stg.Delete(ri.HMACPath(vers.Filename))
				}
			}

			// Update the index.
			if err := ri.AddVersion(cmd.appVersion, repoFilename, !cmd.disabled, onDelete); err != nil {
				cmd.result.AppendError(err)
				return cmd.result
			}

			// Write the index back.
			if err := setRepoIndex(stg, ri); err != nil {
				cmd.result.AppendError(err)
			}
		} else {
			cmd.result.AppendError(err)
			return cmd.result
		}
	} else {
		cmd.result.AppendError(err)
	}

	return cmd.result
}

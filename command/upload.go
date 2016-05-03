package command

import (
	"flag"
	"os"
	"path"

	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/storage"
)

// pulldeploy upload -app=<app> -version=<version> [-disabled] <file>
type Upload struct {
	el         *ErrorList
	pdcfg      pdconfig.PDConfig
	appName    string
	appVersion string
	disabled   bool
	filename   string
}

func (cmd *Upload) CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList {

	var appName, appVersion string
	var disabled bool
	cmd.el = NewErrorList(cmdName)
	cmd.pdcfg = pdcfg

	cmdFlags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmdFlags.StringVar(&appName, "app", "", "name of the application")
	cmdFlags.StringVar(&appVersion, "version", "", "version of the application being uploaded")
	cmdFlags.BoolVar(&disabled, "disabled", false, "upload in disabled state")
	cmdFlags.Parse(osArgs)

	if appName == "" {
		cmd.el.Errorf("app is a mandatory argument")
	} else {
		cmd.appName = appName
	}

	if appVersion == "" {
		cmd.el.Errorf("version is a mandatory argument")
	} else {
		cmd.appVersion = appVersion
	}

	cmd.disabled = disabled

	if len(cmdFlags.Args()) < 1 {
		cmd.el.Errorf("filename is a mandatory argument")
	} else if len(cmdFlags.Args()) > 1 {
		cmd.el.Errorf("only one filename may be specified")
	} else {
		cmd.filename = cmdFlags.Args()[0]
	}

	return cmd.el
}

func (cmd *Upload) Exec() *ErrorList {

	// Ensure the app definition exists.
	if _, err := cmd.pdcfg.GetAppConfig(cmd.appName); err != nil {
		cmd.el.Append(err)
		return cmd.el
	}

	// Get access to the repo storage.
	stgcfg := cmd.pdcfg.GetStorageConfig()
	stg, err := storage.NewStorage(stgcfg.Type, stgcfg.Params)
	if err != nil {
		cmd.el.Append(err)
		return cmd.el
	}

	// Retrieve the repository index.
	if ri, err := getRepoIndex(stg, cmd.appName); err == nil {

		// Open the artifact to be uploaded.
		if fh, err := os.Open(cmd.filename); err == nil {
			defer fh.Close()

			// Determine the content length.
			fi, err := fh.Stat()
			if err != nil {
				cmd.el.Append(err)
				return cmd.el
			}

			// Write the artifact to the repo.
			repoFilename := ri.ArtifactFilename(cmd.appVersion, path.Base(cmd.filename))
			repoPath := ri.ArtifactPath(repoFilename)
			if err := stg.PutReader(repoPath, fh, fi.Size()); err != nil {
				cmd.el.Append(err)
				return cmd.el
			}

			// Update the index.
			if err := ri.AddVersion(cmd.appVersion, repoFilename, !cmd.disabled); err != nil {
				cmd.el.Append(err)
				return cmd.el
			}

			// Write the index back.
			if err := setRepoIndex(stg, ri); err != nil {
				cmd.el.Append(err)
			}
		} else {
			cmd.el.Append(err)
			return cmd.el
		}
	} else {
		cmd.el.Append(err)
	}

	return cmd.el
}

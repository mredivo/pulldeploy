package deployment

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// Provide a dummy configuration.
type mypdConfig struct{}

var pdcfg *mypdConfig

func (p *mypdConfig) GetArtifactConfig(artifactType string) (*pdconfig.ArtifactConfig, error) {
	var ac pdconfig.ArtifactConfig
	ac.Extension = "tar.gz"
	switch runtime.GOOS {
	case "darwin":
		ac.Extract.Cmd = "/usr/bin/tar"
	default:
		ac.Extract.Cmd = "/bin/tar"
	}
	ac.Extract.Args = []string{"zxf", "#ARTIFACTPATH#", "-C", "#VERSIONDIR#"}
	return &ac, nil
}

func (p *mypdConfig) GetAppConfig(appName string) (*pdconfig.AppConfig, error) {
	var appConfig pdconfig.AppConfig
	return &appConfig, nil
}

func (p *mypdConfig) GetAppList() map[string]*pdconfig.AppConfig {
	return make(map[string]*pdconfig.AppConfig)
}

func (p *mypdConfig) GetLogLevel() string {
	return "debug"
}

func (p *mypdConfig) GetSignallerConfig() *pdconfig.SignallerConfig {
	var signaller pdconfig.SignallerConfig
	return &signaller
}

func (p *mypdConfig) GetStorageConfig() *pdconfig.StorageConfig {
	var sc pdconfig.StorageConfig
	return &sc
}

func (p *mypdConfig) GetVersionInfo() *pdconfig.VersionInfo {
	var versionInfo pdconfig.VersionInfo
	return &versionInfo
}

func (p *mypdConfig) RefreshAppList() []error {
	var errs []error = make([]error, 0)
	return errs
}

func TestDeploymentOperations(t *testing.T) {

	const TESTAPP = "stubapp"

	secret := "the quick brown fox jumps over the lazy dog"

	badHMAC := []byte("Invalid HMAC value for testing")
	goodHMAC := []byte("\x13\xb4\x8c\\\x8a\xb9-]\xb5\xdbʱA ̙\x83\xd8.8\x94\x06\"\xb13\xc5\xf3\xf7\xf8\x16\xde\x02")

	os.RemoveAll("../data/client/" + TESTAPP)

	// Test the failure modes.
	appcfg := &pdconfig.AppConfig{Secret: secret, ArtifactType: "tar.gz"}
	if _, err := New(TESTAPP, pdcfg, appcfg); err == nil {
		t.Errorf("Deployment initialization succeeded with missing root dir")
	} else {
		fmt.Println(err.Error())
	}

	appcfg = &pdconfig.AppConfig{Secret: secret, ArtifactType: "tar.gz", BaseDir: "../data/nosuchdir"}
	if _, err := New("", pdcfg, appcfg); err == nil {
		t.Errorf("Deployment initialization succeeded with missing appname")
	} else {
		fmt.Println(err.Error())
	}

	appcfg = &pdconfig.AppConfig{Secret: secret, ArtifactType: "tar.gz", BaseDir: "/"}
	if _, err := New(TESTAPP, pdcfg, appcfg); err == nil {
		t.Errorf("Deployment initialization succeeded with root dir \"/\"")
	} else {
		fmt.Println(err.Error())
	}

	appcfg = &pdconfig.AppConfig{Secret: secret, ArtifactType: "tar.gz", BaseDir: "/foo"}
	if _, err := New(TESTAPP, pdcfg, appcfg); err == nil {
		t.Errorf("Deployment initialization succeeded with root path too short")
	} else {
		fmt.Println(err.Error())
	}

	appcfg = &pdconfig.AppConfig{Secret: secret, ArtifactType: "tar.gz", BaseDir: "../data/nosuchdir"}
	if _, err := New(TESTAPP, pdcfg, appcfg); err == nil {
		t.Errorf("Deployment initialization succeeded with bad root dir")
	} else {
		fmt.Println(err.Error())
	}

	// Create a Deployment for further testing.
	appcfg = &pdconfig.AppConfig{Secret: secret, ArtifactType: "tar.gz", BaseDir: "../data/client"}
	dep, err := New(TESTAPP, pdcfg, appcfg)
	if err != nil {
		t.Errorf("Deployment initialization failed: %s", err.Error())
	}

	// Write some bytes to an artifact.
	if fp, err := os.Open("../data/testdata/stubapp.tar.gz"); err == nil {
		if err := dep.WriteArtifact("1.0.3", fp); err != nil {
			t.Errorf("WriteArtifact failed: %s", err.Error())
			fp.Close()
		}
	} else {
		t.Errorf("Could not open test data file for reading: %s", err.Error())
	}

	// Writing the same artifact again should fail.
	if fp, err := os.Open("../data/testdata/stubapp.tar.gz"); err == nil {
		defer fp.Close()
		if err := dep.WriteArtifact("1.0.3", fp); err == nil {
			t.Errorf("WriteArtifact should have failed as DUPLICATE")
		}
	} else {
		t.Errorf("Could not open test data file for reading: %s", err.Error())
	}

	// Write an invalid HMAC.
	if err := dep.WriteHMAC("1.0.3", badHMAC); err != nil {
		t.Errorf("WriteHMAC failed: %s", err.Error())
	}

	// Validate the HMAC.
	if err := dep.CheckHMAC("1.0.3"); err != nil {
		fmt.Printf("CheckHMAC failed: %s\n", err.Error())
	} else {
		t.Errorf("CheckHMAC succeeded, but should not have\n")
	}

	// Write a valid HMAC.
	if err := dep.WriteHMAC("1.0.3", goodHMAC); err != nil {
		t.Errorf("WriteHMAC failed: %s", err.Error())
	}

	// Validate the HMAC.
	if err := dep.CheckHMAC("1.0.3"); err != nil {
		t.Errorf("CheckHMAC failed: %s\n", err.Error())
	} else {
		fmt.Printf("CheckHMAC succeeded\n")
	}

	// Extract the artifact into the release directory.
	if err := dep.Extract("1.0.3"); err != nil {
		t.Errorf("Extract failed: %s", err.Error())
	}

	// There should be no current link yet.
	if current := dep.GetCurrentLink(); current != "" {
		t.Errorf("Current link set to %q; should be empty", current)
	}

	// Make it current.
	if err := dep.Link("1.0.3"); err != nil {
		t.Errorf("Link failed: %s", err.Error())
	}

	// Current link should be what we set it.
	if current := dep.GetCurrentLink(); current != "1.0.3" {
		t.Errorf("Current link should be %q; found  %q", "1.0.3", current)
	}

	// Publish two more versions, so we can list them.
	if fp, err := os.Open("../data/testdata/stubapp.tar.gz"); err == nil {
		if err := dep.WriteArtifact("1.0.4", fp); err != nil {
			t.Errorf("WriteArtifact failed: %s", err.Error())
			fp.Close()
		}
	} else {
		t.Errorf("Could not open test data file for reading: %s", err.Error())
	}
	if err := dep.Extract("1.0.4"); err != nil {
		t.Errorf("Extract failed: %s", err.Error())
	}

	if fp, err := os.Open("../data/testdata/stubapp.tar.gz"); err == nil {
		if err := dep.WriteArtifact("1.1.0", fp); err != nil {
			t.Errorf("WriteArtifact failed: %s", err.Error())
			fp.Close()
		}
	} else {
		t.Errorf("Could not open test data file for reading: %s", err.Error())
	}
	if err := dep.Extract("1.1.0"); err != nil {
		t.Errorf("Extract failed: %s", err.Error())
	}

	// List the versions available.
	if versionList := dep.ListVersions(); len(versionList) != 3 {
		t.Errorf("ListVersions failed: expected 3, got %v", versionList)
	} else {
		fmt.Println(versionList)
	}

	// Link a different version.
	if err := dep.Link("1.0.4"); err != nil {
		t.Errorf("Link failed: %s", err.Error())
	}

	// Current link should be what we set it.
	if current := dep.GetCurrentLink(); current != "1.0.4" {
		t.Errorf("Current link should be %q; found  %q", "1.0.4", current)
	}

	// Link a bogus version.
	if err := dep.Link("9.9.9"); err == nil {
		t.Errorf("Link non-existent version did not fail")
	}

	// Attempt to remove the current version.
	if err := dep.Remove("1.0.4"); err == nil {
		t.Errorf("Remove current version did not fail")
	}

	// Remove the previous version.
	if err := dep.Remove("1.0.3"); err != nil {
		t.Errorf("Remove previous version failed: %s", err.Error())
	}
}

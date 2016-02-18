package deployment

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
)

func TestDeploymentOperations(t *testing.T) {

	const TESTAPP = "stubapp"

	hmacKey := []byte("the quick brown fox jumps over the lazy dog")
	hmacCalculator := hmac.New(sha256.New, hmacKey)

	badHMAC := []byte("Invalid HMAC value for testing")
	goodHMAC := []byte("\x13\xb4\x8c\\\x8a\xb9-]\xb5\xdbʱA ̙\x83\xd8.8\x94\x06\"\xb13\xc5\xf3\xf7\xf8\x16\xde\x02")

	os.RemoveAll("../data/client/" + TESTAPP)

	dep := new(Deployment)

	// Test the failure modes.
	if err := dep.Init(TESTAPP, "tar.gz", "", 0, 0); err == nil {
		t.Errorf("Deployment initialization succeeded with missing root dir")
	} else {
		fmt.Println(err.Error())
	}
	if err := dep.Init("", "tar.gz", "../data/nosuchdir", 0, 0); err == nil {
		t.Errorf("Deployment initialization succeeded with missing base dir")
	} else {
		fmt.Println(err.Error())
	}
	if err := dep.Init(TESTAPP, "tar.gz", "/", 0, 0); err == nil {
		t.Errorf("Deployment initialization succeeded with root dir \"/\"")
	} else {
		fmt.Println(err.Error())
	}
	if err := dep.Init(TESTAPP, "tar.gz", "/foo", 0, 0); err == nil {
		t.Errorf("Deployment initialization succeeded with root path too short")
	} else {
		fmt.Println(err.Error())
	}
	if err := dep.Init(TESTAPP, "tar.gz", "../data/nosuchdir", 0, 0); err == nil {
		t.Errorf("Deployment initialization succeeded with bad root dir")
	} else {
		fmt.Println(err.Error())
	}

	// Create a Deployment for further testing.
	if err := dep.Init(TESTAPP, "tar.gz", "../data/client", 1001, 1001); err != nil {
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
	if err := dep.WriteSignature("1.0.3", badHMAC); err != nil {
		t.Errorf("WriteSignature failed: %s", err.Error())
	}

	// Validate the signature.
	if err := dep.CheckSignature("1.0.3", hmacCalculator); err != nil {
		fmt.Printf("CheckSignature failed: %s\n", err.Error())
	} else {
		t.Errorf("CheckSignature succeeded, but should not have\n")
	}

	// Write a valid HMAC.
	if err := dep.WriteSignature("1.0.3", goodHMAC); err != nil {
		t.Errorf("WriteSignature failed: %s", err.Error())
	}

	// Validate the signature.
	if err := dep.CheckSignature("1.0.3", hmacCalculator); err != nil {
		t.Errorf("CheckSignature failed: %s\n", err.Error())
	} else {
		fmt.Printf("CheckSignature succeeded\n")
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

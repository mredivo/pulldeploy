package repo

import (
	"fmt"
	"testing"
)

func TestIndex(t *testing.T) {

	ri := NewIndex("Example_App")
	if ri == nil {
		t.Errorf("Index creation failed")
	}
	if ri.appName != "example_app" {
		t.Errorf("Index appName not correctly set")
	}
	if ri.Canary != 0 {
		t.Errorf("Index Canary not correctly set")
	}
	if len(ri.Versions) != 0 {
		t.Errorf("Index Versions[] not correctly set")
	}
	if len(ri.Envs) != 0 {
		t.Errorf("Index Envs[] not correctly set")
	}
}

func TestIndexEnvs(t *testing.T) {

	envName := "staging"
	ri := NewIndex("Example_App")

	// Test GetEnv/RmEnv while the list is empty.
	if _, err := ri.GetEnv(envName); err == nil {
		t.Errorf("Index GetEnv should have failed")
	}

	if err := ri.RmEnv(envName); err == nil {
		t.Errorf("Index RmEnv should have failed")
	}

	// Add an environment.
	if err := ri.AddEnv(envName); err != nil {
		t.Errorf("Index AddEnv failed: %s", err.Error())
	}
	if err := ri.AddEnv(envName); err == nil {
		t.Errorf("Index AddEnv should have failed as duplicate")
	}
	if len(ri.Envs) != 1 {
		t.Errorf("Index len(Envs[]) is %d, should be 1", len(ri.Envs))
	}

	// Add another environment, to make sure following tests work in that case.
	if err := ri.AddEnv("production"); err != nil {
		t.Errorf("Index AddEnv failed: %s", err.Error())
	}
	if len(ri.Envs) != 2 {
		t.Errorf("Index len(Envs[]) is %d, should be 2", len(ri.Envs))
	}

	// Get the environment.
	if _, err := ri.GetEnv(envName); err != nil {
		t.Errorf("Index GetEnv failed: %s", err.Error())
	}
	if len(ri.Envs) != 2 {
		t.Errorf("Index len(Envs[]) is %d, should be 2", len(ri.Envs))
	}
	if _, err := ri.GetEnv(envName); err != nil {
		t.Errorf("Index GetEnv failed: %s", err.Error())
	}
	if len(ri.Envs) != 2 {
		t.Errorf("Index len(Envs[]) is %d, should be 2", len(ri.Envs))
	}

	// Update the environment, and confirm that the change sticks.
	env, _ := ri.GetEnv(envName)
	if env.Preview != "" {
		t.Errorf("Index unexpected Preview; should be blank")
	}
	env.Preview = "1.0.2"
	if err := ri.SetEnv(envName, env); err != nil {
		t.Errorf("Index SetEnv failed: %s", err.Error())
	}
	env, _ = ri.GetEnv(envName)
	if env.Preview != "1.0.2" {
		t.Errorf("Index unexpected Preview; should be \"1.0.2\"")
	}

	// Remove the environment.
	if err := ri.RmEnv(envName); err != nil {
		t.Errorf("Index RmEnv failed: %s", err.Error())
	}
	if len(ri.Envs) != 1 {
		t.Errorf("Index len(Envs[]) is %d, should be 1", len(ri.Envs))
	}
	if err := ri.RmEnv(envName); err == nil {
		t.Errorf("Index RmEnv should have failed")
	}
	if len(ri.Envs) != 1 {
		t.Errorf("Index len(Envs[]) is %d, should be 1", len(ri.Envs))
	}
}

func TestIndexVersions(t *testing.T) {

	versionName := "1.0.1"
	ri := NewIndex("Example_App")

	// Ensure that minimum number of versions to keep is not zero.
	if err := ri.AddEnv("dummy"); err != nil {
		t.Errorf("Index AddEnv failed: %s", err.Error())
	}

	onDelete := func(versionName string) {
		fmt.Printf("Version %q removed from list\n", versionName)
	}

	// Test GetVersion/RmVersion while the list is empty.
	if _, err := ri.GetVersion(versionName); err == nil {
		t.Errorf("Index GetVersion should have failed")
	}

	if err := ri.RmVersion(versionName); err == nil {
		t.Errorf("Index RmVersion should have failed")
	}

	// Add a version.
	if err := ri.AddVersion(versionName, "foo.tar.gz", true, onDelete); err != nil {
		t.Errorf("Index AddVersion failed: %s", err.Error())
	}
	if err := ri.AddVersion(versionName, "foo.tar.gz", true, onDelete); err == nil {
		t.Errorf("Index AddVersion should have failed as duplicate")
	}
	if len(ri.Versions) != 1 {
		t.Errorf("Index len(Versions[]) is %d, should be 1", len(ri.Versions))
	}

	// Add another version, to make sure following tests work in that case.
	if err := ri.AddVersion("1.0.2", "foo.tar.gz", true, onDelete); err != nil {
		t.Errorf("Index AddVersion failed: %s", err.Error())
	}
	if len(ri.Versions) != 2 {
		t.Errorf("Index len(Versions[]) is %d, should be 2", len(ri.Versions))
	}

	// Get the version.
	if _, err := ri.GetVersion(versionName); err != nil {
		t.Errorf("Index GetVersion failed: %s", err.Error())
	}
	if len(ri.Versions) != 2 {
		t.Errorf("Index len(Versions[]) is %d, should be 2", len(ri.Versions))
	}
	if _, err := ri.GetVersion(versionName); err != nil {
		t.Errorf("Index GetVersion failed: %s", err.Error())
	}
	if len(ri.Versions) != 2 {
		t.Errorf("Index len(Versions[]) is %d, should be 2", len(ri.Versions))
	}

	// Update the version, and confirm that the change sticks.
	vers, _ := ri.GetVersion(versionName)
	if !vers.Enabled {
		t.Errorf("Index unexpected Enabled state; should be enabled")
	}
	vers.Disable()
	if err := ri.SetVersion(versionName, vers); err != nil {
		t.Errorf("Index SetVersion failed: %s", err.Error())
	}
	vers, _ = ri.GetVersion(versionName)
	if vers.Enabled {
		t.Errorf("Index unexpected Enabled state; should be disabled")
	}

	// Remove the version.
	if err := ri.RmVersion(versionName); err != nil {
		t.Errorf("Index RmVersion failed: %s", err.Error())
	}
	if len(ri.Versions) != 1 {
		t.Errorf("Index len(Versions[]) is %d, should be 1", len(ri.Versions))
	}
	if err := ri.RmVersion(versionName); err == nil {
		t.Errorf("Index RmVersion should have failed")
	}
	if len(ri.Versions) != 1 {
		t.Errorf("Index len(Versions[]) is %d, should be 1", len(ri.Versions))
	}
}

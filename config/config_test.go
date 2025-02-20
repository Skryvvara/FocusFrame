package config

import (
	"fmt"
	"os"
	"path"
	"testing"
)

func getPaths() (string, string, string) {
	configPath := getConfigPath()
	basePath := path.Dir(configPath)
	tmpPath := path.Join(os.Getenv("APPDATA"), "_FocusFrame")

	return configPath, basePath, tmpPath
}

func setup() (bool, error) {
	_, basePath, tmpPath := getPaths()

	if _, err := os.Stat(basePath); err == nil {
		if err := os.Rename(basePath, tmpPath); err != nil {
			return false, fmt.Errorf("Could not move existing config dir from '%s' to '%s' with error: %v", basePath, tmpPath, err)
		}
		return true, nil
	}

	return false, nil
}

func cleanup(movedDir bool) error {
	_, basePath, tmpPath := getPaths()

	if err := os.RemoveAll(basePath); err != nil {
		return fmt.Errorf("Could not remove newly created config dir at '%s' with error: %v", basePath, err)
	}

	if movedDir {
		if err := os.Rename(tmpPath, basePath); err != nil {
			return fmt.Errorf("Could not move tmp config dir back from '%s' to '%s' with error: %v", tmpPath, basePath, err)
		}
	}

	return nil
}

func TestAssertPath(t *testing.T) {
	testPath := path.Join("./", "very", "very", "nested", "test", "path")
	if err := assertPath(testPath); err != nil {
		t.Fatalf("Failed to assert path '%s' with error: %v", testPath, err)
	}
	if _, err := os.Stat(testPath); err != nil {
		t.Fatalf("Stat failed for test path '%s' with error: %v", testPath, err)
	}
	if err := os.RemoveAll(path.Join("./", "very")); err != nil {
		t.Fatalf("Failed to remove test path '%s' with error: %v", testPath, err)
	}
}

func TestInitialize(t *testing.T) {
	movedDir, err := setup()
	if err != nil {
		t.Fatalf("%v", err)
	}

	Initialize()

	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Failed to create config file or dir at '%s' with error: %v", configPath, err)
	}

	if err := cleanup(movedDir); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestAddApplication(t *testing.T) {
	movedDir, err := setup()
	if err != nil {
		t.Fatalf("%v", err)
	}

	Initialize()

	testApp := "TestApp.exe"
	AddApplication(testApp)
	Config = Type{}

	loadConfig()

	hit := false
	for _, v := range Config.ManagedApps {
		if v.Executable == testApp {
			hit = true
		}
	}

	if !hit {
		t.Fatalf("Test application '%s' is not present after re-loading config file.", testApp)
	}

	if err := cleanup(movedDir); err != nil {
		t.Fatalf("%v", err)
	}
}

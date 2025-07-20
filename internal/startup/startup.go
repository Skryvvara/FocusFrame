package startup

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

// Path to the Run-Key
const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

// The name for the registry value
const appName = "FocusFrame"

// Enable adds the app to the Run registry key
//
// Returns an error if the entry could not be created.
func Enable() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	key, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.SetStringValue(appName, `"`+exePath+`"`)
}

// Disable removes the app from the Run registry key
//
// Returns an error if the entry could not be removed.
func Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.DeleteValue(appName)
}

// IsEnabled checks if the app is in the Run registry key
//
// Returns a bool if the entry could be found, otherwise an error is returned.
func IsEnabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false, err
	}
	defer key.Close()

	val, _, err := key.GetStringValue(appName)
	if err == registry.ErrNotExist {
		return false, nil
	} else if err != nil {
		return false, err
	}

	exePath, err := os.Executable()
	if err != nil {
		return false, err
	}

	return val == `"`+exePath+`"`, nil
}

package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"path"
	"sync"
)

type WindowSettings struct {
	Width   int `toml:"width"`
	Height  int `toml:"height"`
	OffsetX int `toml:"offsetX"`
	OffsetY int `toml:"offsetY"`
}

type ManagedApp struct {
	Executable string         `toml:"executable"`
	Dimensions WindowSettings `toml:"dimensions"`
}

type Type struct {
	Global struct {
		Width   int `toml:"width"`
		Height  int `toml:"height"`
		OffsetX int `toml:"offsetX"`
		OffsetY int `toml:"offsetY"`
	} `toml:"global"`
	ManagedApps     map[string]ManagedApp `toml:"managed_apps"`
	managedAppsLock sync.Mutex
}

var Config Type
var configPath = path.Join(".", "config.toml")

// Initialize loads the configuration from file into the Config struct.
func Initialize() {
	Config.ManagedApps = make(map[string]ManagedApp)

	if _, err := toml.DecodeFile(configPath, &Config); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
	}
}

// AddApplication adds the given executable to the config and tries to write the changes to the config file.
func AddApplication(executable string) {
	Config.managedAppsLock.Lock()
	defer Config.managedAppsLock.Unlock()

	// Get global window settings for the dimensions
	dimensions := getGlobalWindowSettings()

	// Create the ManagedApp instance
	app := ManagedApp{
		Executable: executable,
		Dimensions: dimensions,
	}

	// Add the app to the ManagedApps map
	Config.ManagedApps[executable] = app

	// Save the updated configuration
	if err := saveConfig(); err != nil {
		fmt.Printf("Failed to save config after adding application: %s\n", err)
	}
}

// RemoveApplication removes the given executable from the config and tries to write the changes to the config file.
func RemoveApplication(executable string) {
	Config.managedAppsLock.Lock()
	defer Config.managedAppsLock.Unlock()

	// Remove the app from the ManagedApps map
	delete(Config.ManagedApps, executable)

	// Save the updated configuration
	if err := saveConfig(); err != nil {
		fmt.Printf("Failed to save config after removing application: %s\n", err)
	}
}

// saveConfig tries to write the current configuration to file and returns an error if it fails.
func saveConfig() error {
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := toml.NewEncoder(file).Encode(Config); err != nil {
		return err
	}
	return nil
}

// IsValid checks if the window dimensions are valid.
func (ws WindowSettings) IsValid() bool {
	// Check that width and height are valid integers and greater or equal 0
	if ws.Width < 0 || ws.Height < 0 {
		return false
	}

	// Additional checks can be added here as needed
	return true
}

// GetWindowSettings returns the WindowSettings specific to a managed application or the global WindowSettings if no
// specific or invalid settings were found.
func GetWindowSettings(executable string) WindowSettings {
	for _, app := range Config.ManagedApps {
		if app.Executable == executable {
			if app.Dimensions.IsValid() {
				return app.Dimensions
			}
		}
	}
	return getGlobalWindowSettings()
}

// getGlobalWindowSettings returns a WindowSettings struct holding the global configuration.
func getGlobalWindowSettings() WindowSettings {
	return WindowSettings{
		Width:   Config.Global.Width,
		Height:  Config.Global.Height,
		OffsetX: Config.Global.OffsetX,
		OffsetY: Config.Global.OffsetY,
	}
}

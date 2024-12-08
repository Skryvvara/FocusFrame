package config

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/creasty/defaults"
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
		Width   int `toml:"width" default:"1920"`
		Height  int `toml:"height" default:"1090"`
		OffsetX int `toml:"offsetX" default:"0"`
		OffsetY int `toml:"offsetY" default:"0"`
		Hotkey  int `toml:"hotkey" default:"115"`
	} `toml:"global"`
	ManagedApps     map[string]ManagedApp `toml:"managed_apps"`
	managedAppsLock sync.Mutex
}

var Config Type
var configPath string

// Initialize loads the configuration from file into the Config struct.
func Initialize() {
	Config.ManagedApps = make(map[string]ManagedApp)

	configPath = getConfigPath()

	defaults.Set(&Config)

	// try to create empty config file if file doesn't exist
	if _, err := os.Stat(configPath); err != nil {
		if err = saveConfig(); err != nil {
			log.Fatal(err)
		}
	}

	if _, err := toml.DecodeFile(configPath, &Config); err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}
}

// getConfigPath returns the configuration path according to the runtime os (including the filename).
func getConfigPath() string {
	filename := "config.toml"

	if runtime.GOOS == "windows" {
		return path.Join(os.Getenv("APPDATA"), "FocusFrame", filename)
	} else {
		return path.Join(os.Getenv("HOME"), ".config", "FocusFrame", filename)
	}
}

// OpenConfigPath tried to issue a command based on the runtime os to reveal the configuration file in a file explorer.
func OpenConfigPath() error {
	configDir := filepath.Dir(configPath)
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("explorer", configDir)
		if err := cmd.Start(); err != nil {
			return err
		}
	case "darwin":
		cmd := exec.Command("open", configDir)
		if err := cmd.Start(); err != nil {
			return err
		}
	case "linux":
		cmd := exec.Command("xdg-open", configDir)
		if err := cmd.Start(); err != nil {
			return err
		}
	default:
		log.Fatalf("Unsupported platform: %s", runtime.GOOS)
	}
	return nil
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

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

var Version string

type WindowSettings struct {
	Width   int `toml:"width"`
	Height  int `toml:"height"`
	OffsetX int `toml:"offsetX"`
	OffsetY int `toml:"offsetY"`
	Delay   int `toml:"delay"`
}

type ManagedApp struct {
	Executable   string         `toml:"executable"`
	FriendlyName string         `toml:"friendly_name"`
	Dimensions   WindowSettings `toml:"dimensions"`
}

type Type struct {
	Global struct {
		Width     int  `toml:"width" default:"1920"`
		Height    int  `toml:"height" default:"1090"`
		OffsetX   int  `toml:"offsetX" default:"0"`
		OffsetY   int  `toml:"offsetY" default:"0"`
		Delay     int  `toml:"delay" default:"0"`
		Hotkey    int  `toml:"hotkey" default:"115"`
		DarkTheme bool `toml:"dark_theme" default:"false"`
	} `toml:"global"`
	ManagedApps map[string]ManagedApp `toml:"managed_apps"`
}

var ManagedAppsLock sync.Mutex
var Config Type
var configPath string

// Initialize loads the configuration from file into the Config struct.
func Initialize() {
	Config.ManagedApps = make(map[string]ManagedApp)

	configPath = getConfigPath()

	if err := assertPath(path.Dir(configPath)); err != nil {
		log.Fatal(err)
	}

	defaults.Set(&Config)

	// try to create empty config file if file doesn't exist
	if _, err := os.Stat(configPath); err != nil {
		if err = SaveConfig(); err != nil {
			log.Fatal(err)
		}
	}

	if err := loadConfig(); err != nil {
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

// assertPath checks if the given path exists, alternatively it tries to create all missing directories of the path.
//
// returns nil if the path exists or could be created, returns an error if creation of the path failed.
func assertPath(path string) error {
	if _, err := os.Stat(path); err != nil {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
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
	ManagedAppsLock.Lock()
	defer ManagedAppsLock.Unlock()

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
	if err := SaveConfig(); err != nil {
		fmt.Printf("Failed to save config after adding application: %s\n", err)
	}
}

// RemoveApplication removes the given executable from the config and tries to write the changes to the config file.
func RemoveApplication(executable string) {
	ManagedAppsLock.Lock()
	defer ManagedAppsLock.Unlock()

	// Remove the app from the ManagedApps map
	delete(Config.ManagedApps, executable)

	// Save the updated configuration
	if err := SaveConfig(); err != nil {
		fmt.Printf("Failed to save config after removing application: %s\n", err)
	}
}

// SaveConfig tries to write the current configuration to file and returns an error if it fails.
func SaveConfig() error {
	if len(configPath) <= 0 {
		return fmt.Errorf("configPath is not set")
	}

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = toml.NewEncoder(file).Encode(Config)
	return err
}

// loadConfig tries to read the configruation from the config file and returns an error if it fails.
func loadConfig() error {
	if len(configPath) <= 0 {
		return fmt.Errorf("configPath is not set")
	}

	_, err := toml.DecodeFile(configPath, &Config)
	return err
}

// IsValid checks if the window dimensions are valid.
func (ws WindowSettings) IsValid() bool {
	// Check that width and height are valid integers and greater or equal 0
	if ws.Width < 0 || ws.Height < 0 {
		return false
	}

	if ws.Delay < 0 {
		return false
	}

	// Additional checks can be added here as needed
	return true
}

// GetWindowSettings returns the WindowSettings specific to a managed application or the global WindowSettings if no
// specific or invalid settings were found.
func GetWindowSettings(executable string) WindowSettings {
	if app, ok := Config.ManagedApps[executable]; ok {
		return app.Dimensions
	}
	return getGlobalWindowSettings()
}

// getGlobalWindowSettings returns a WindowSettings struct holding the global configuration.
func getGlobalWindowSettings() WindowSettings {
	return GetWindowSettingsFromStruct(Config)
}

func GetWindowSettingsFromStruct(config Type) WindowSettings {
	return WindowSettings{
		Width:   config.Global.Width,
		Height:  config.Global.Height,
		OffsetX: config.Global.OffsetX,
		OffsetY: config.Global.OffsetY,
		Delay:   config.Global.Delay,
	}
}

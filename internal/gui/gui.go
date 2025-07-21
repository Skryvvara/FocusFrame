package gui

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/skryvvara/focusframe/config"
	webview "github.com/webview/webview_go"
)

const (
	REPO_URL = "https://github.com/skryvvara/focusframe"
)

//go:embed index.html
var viewFS embed.FS

// ShowGUI initializes and runs the GUI using the native webview
func ShowGUI() {
	htmlBytes, err := viewFS.ReadFile("index.html")
	if err != nil {
		log.Println("Failed to read embedded GUI HTML:", err)
		return
	}
	encoded := base64.StdEncoding.EncodeToString(htmlBytes)
	dataURI := "data:text/html;base64," + encoded

	w := webview.New(true)
	defer w.Destroy()

	w.Bind("getGlobalConfig", func() any {
		return config.Config.Global
	})

	w.Bind("getManagedApps", bindGetManagedApps)
	w.Bind("saveGlobalConfigChanges", bindSaveGlobalConfigChanges)
	w.Bind("saveAppChanges", bindSaveAppChanges)

	w.SetTitle(fmt.Sprintf("FocusFrame %s", config.Version))
	w.SetSize(500, 475, webview.HintNone)
	w.Navigate(dataURI)

	w.Run()
}

// bindGetManagedApps returns the list of managed applications as a JSON string.
func bindGetManagedApps() string {
	data, err := json.Marshal(config.Config.ManagedApps)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// bindSaveGlobalConfigChanges updates the global configuration using data passed from the GUI.
func bindSaveGlobalConfigChanges(data map[string]interface{}) {
	cfg := &config.Config.Global
	if v, ok := data["Width"].(float64); ok {
		cfg.Width = int(v)
	}
	if v, ok := data["Height"].(float64); ok {
		cfg.Height = int(v)
	}
	if v, ok := data["OffsetX"].(float64); ok {
		cfg.OffsetX = int(v)
	}
	if v, ok := data["OffsetY"].(float64); ok {
		cfg.OffsetY = int(v)
	}
	if v, ok := data["Delay"].(float64); ok {
		cfg.Delay = int(v)
	}
	if v, ok := data["Hotkey"].(float64); ok {
		cfg.Hotkey = int(v)
	}

	err := config.SaveConfig()
	if err != nil {
		log.Println("Failed to save config:", err)
	}
}

// bindSaveAppChanges updates a managed application's settings using data passed from the GUI.
func bindSaveAppChanges(data map[string]interface{}) {
	bytes, err := json.Marshal(data)
	if err != nil {
		log.Println("Failed to marshal data:", err)
		return
	}

	var newAppSettings config.ManagedApp
	err = json.Unmarshal(bytes, &newAppSettings)
	if err != nil {
		log.Println("Failed to unmarshal into struct:", err)
		return
	}

	if newAppSettings.Executable == "" {
		log.Print("Executable cannot be empty")
		return
	}

	if !newAppSettings.Dimensions.IsValid() {
		log.Println("Window dimension settings are invalid")
		return
	}

	if _, ok := config.Config.ManagedApps[newAppSettings.Executable]; !ok {
		log.Println("Failed to update")
		return
	}

	config.Config.ManagedApps[newAppSettings.Executable] = newAppSettings
	err = config.SaveConfig()
	if err != nil {
		log.Println("Failed to save config:", err)
	}
}

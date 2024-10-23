package main

import (
	"embed"
	"fmt"
	"github.com/getlantern/systray"
	"github.com/skryvvara/focusframe/config"
	"github.com/skryvvara/focusframe/input"
	"github.com/skryvvara/focusframe/window"
	"log"
)

var Version = "vX.Y.Z" // this is set during build time

//go:embed monitor.ico
var iconFS embed.FS

func main() {
	config.Initialize()

	go window.AddAppOnKeyPress(input.VK_F4)
	go window.WatchForegroundWindowChange()

	systray.Run(onReady, onExit)
}

// onReady setup systray
func onReady() {
	iconData, err := iconFS.ReadFile("monitor.ico")
	if err != nil {
		log.Fatal("Error reading icon: ", err)
	}

	systray.SetIcon(iconData)
	systray.SetTitle("FocusFrame")
	systray.SetTooltip(fmt.Sprintf("FocusFrame Version: %s", Version))

	systray.AddSeparator()

	mRunOnStartupCh := systray.AddMenuItem("Run on Startup", "Run the application when starting the PC (Not implemented)")

	systray.AddSeparator()

	mQuitCh := systray.AddMenuItem("Quit", "Quit the whole app")

	for {
		select {
		case <-mRunOnStartupCh.ClickedCh:
			if mRunOnStartupCh.Checked() {
				mRunOnStartupCh.Uncheck()
			} else {
				mRunOnStartupCh.Check()
			}
		case <-mQuitCh.ClickedCh:
			log.Println("Requesting Exit")
			systray.Quit()
		}
	}
}

// onExit use for cleanup later, currently has no use
func onExit() {
	// clean up here
}

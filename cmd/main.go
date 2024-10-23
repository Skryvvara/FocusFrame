package main

import (
	"embed"
	"fmt"
	"github.com/getlantern/systray"
	"github.com/skryvvara/focusframe/config"
	"github.com/skryvvara/focusframe/input"
	"github.com/skryvvara/focusframe/internal/browser"
	"github.com/skryvvara/focusframe/window"
	"log"
)

var Version = "vX.Y.Z" // this is set during build time

const (
	REPO_URL = "https://github.com/skryvvara/focusframe"
)

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

	mManageApplications := systray.AddMenuItem("Manage Applications", "Manage Applications")

	systray.AddSeparator()

	mWiki := systray.AddMenuItem("Open Wiki", "Open Wiki")
	mForum := systray.AddMenuItem("Open Forum", "Open Forum")
	mGithub := systray.AddMenuItem("Open Github", "Open Github repository")
	mRunOnStartup := systray.AddMenuItem("Run on Startup", "Run the application when starting the PC (Not implemented)")

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	for {
		select {
		case <-mManageApplications.ClickedCh:
			log.Println("Not implemented")
		case <-mWiki.ClickedCh:
			if err := browser.OpenURL(REPO_URL + "/wiki"); err != nil {
				log.Println(err)
			}
		case <-mForum.ClickedCh:
			if err := browser.OpenURL(REPO_URL + "/discussions"); err != nil {
				log.Println(err)
			}
		case <-mGithub.ClickedCh:
			if err := browser.OpenURL(REPO_URL); err != nil {
				log.Println(err)
			}
		case <-mRunOnStartup.ClickedCh:
			if mRunOnStartup.Checked() {
				mRunOnStartup.Uncheck()
			} else {
				mRunOnStartup.Check()
			}
		case <-mQuit.ClickedCh:
			fmt.Println("Requesting Exit")
			systray.Quit()
		}
	}
}

// onExit use for cleanup later, currently has no use
func onExit() {
	// clean up here
}

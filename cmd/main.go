package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/getlantern/systray"
	"github.com/skryvvara/focusframe/config"
	"github.com/skryvvara/focusframe/internal/browser"
	"github.com/skryvvara/focusframe/internal/gui"
	"github.com/skryvvara/focusframe/internal/startup"
	"github.com/skryvvara/focusframe/window"
)

//go:embed monitor.ico
var iconFS embed.FS

func main() {
	config.Initialize()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigs
		systray.Quit()
	}()

	go window.AddAppOnKeyPress(config.Config.Global.Hotkey)
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
	systray.SetTooltip(fmt.Sprintf("FocusFrame Version: %s", config.Version))

	mManageApplications := systray.AddMenuItem("Manage Applications", "Manage Applications")

	systray.AddSeparator()

	mShowConfig := systray.AddMenuItem("Show Configuration", "Show Configuration")
	mReloadConfig := systray.AddMenuItem("Reload Configuration", "Reload Configuration")
	mWiki := systray.AddMenuItem("Open Wiki", "Open Wiki")
	mForum := systray.AddMenuItem("Open Forum", "Open Forum")
	mGithub := systray.AddMenuItem("Open Github", "Open Github repository")

	enabledOnStartup, err := startup.IsEnabled()
	if err != nil {
		log.Fatal(err)
	}
	mRunOnStartup := systray.AddMenuItem("Run on Startup", "Run the application when starting the PC")
	if enabledOnStartup {
		mRunOnStartup.Check()
	}

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	for {
		select {
		case <-mManageApplications.ClickedCh:
			go gui.ShowGUI()
		case <-mShowConfig.ClickedCh:
			if err := config.OpenConfigPath(); err != nil {
				log.Println(err)
			}
		case <-mReloadConfig.ClickedCh:
			config.Initialize()
		case <-mWiki.ClickedCh:
			if err := browser.OpenURL(gui.REPO_URL + "/wiki"); err != nil {
				log.Println(err)
			}
		case <-mForum.ClickedCh:
			if err := browser.OpenURL(gui.REPO_URL + "/discussions"); err != nil {
				log.Println(err)
			}
		case <-mGithub.ClickedCh:
			if err := browser.OpenURL(gui.REPO_URL); err != nil {
				log.Println(err)
			}
		case <-mRunOnStartup.ClickedCh:
			if mRunOnStartup.Checked() {
				err := startup.Disable()
				if err != nil {
					log.Println(err)
				} else {
					mRunOnStartup.Uncheck()
				}
			} else {
				err := startup.Enable()
				if err != nil {
					log.Println(err)
				} else {
					mRunOnStartup.Check()
				}
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

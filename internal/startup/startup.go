package startup

import (
	"errors"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// getStartupPath returns a clean path to the currently logged-in user's startup directory.
//
// e.g. "C:\Users\Me\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup"
func getStartupPath() string {
	userprofile := os.Getenv("USERPROFILE")
	return path.Join(userprofile, "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
}

// getLinkPath returns a clean path to the currently logged-in user's startup directory
// in addition to the shortcut file.
//
// e.g. "C:\Users\Me\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup\FocusFrame.lnk"
func getLinkPath() string {
	executable := filepath.Base(os.Args[0])
	executableName := strings.TrimSuffix(executable, filepath.Ext(executable))
	startupPath := getStartupPath()
	return path.Join(startupPath, executableName+".lnk")
}

// IsEnabled checks if a link to the executable is in the startup directory
// and if the link points to the current executable.
//
// Returns false is the link doesn't exist or the link points to a different executable,
// true if the link exists and points to the correct executable, otherwise returns an error.
func IsEnabled() (bool, error) {
	linkPath := getLinkPath()

	_, err := os.Stat(linkPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		} else {
			return false, err
		}
	}

	targetPath, err := getLinkTarget(linkPath)
	if err != nil {
		return false, err
	}

	if targetPath != os.Args[0] {
		return false, nil
	}

	return true, nil
}

// Enable creates a link in the users startup directory.
//
// Returns an error if it fails to create the shortcut.
func Enable() error {
	isEnabled, err := IsEnabled()
	if err != nil {
		return err
	}

	if isEnabled {
		return nil
	}

	err = createLink(os.Args[0], getLinkPath())
	if err != nil {
		return err
	}

	return nil
}

// Disable deletes the shortcut in the users startup directory.
//
// Returns an error if it fails to delete the shortcut.
func Disable() error {
	isEnabled, err := IsEnabled()
	if err != nil {
		return err
	}

	if !isEnabled {
		return nil
	}

	err = os.Remove(getLinkPath())
	if err != nil {
		return err
	}

	return nil
}

// getLinkTarget fetches the target a shortcut points to.
//
// Returns the target executable's path or an error if it fails to get the target.
func getLinkTarget(linkPath string) (string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	if err != nil {
		return "", err
	}
	defer ole.CoUninitialize()

	oleShellObject, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return "", err
	}
	defer oleShellObject.Release()

	wshell, err := oleShellObject.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return "", err
	}
	defer wshell.Release()

	cs, err := oleutil.CallMethod(wshell, "CreateShortcut", linkPath)
	if err != nil {
		return "", err
	}
	defer cs.Clear()

	idispatch := cs.ToIDispatch()

	targetPathRaw, err := oleutil.GetProperty(idispatch, "TargetPath")
	if err != nil {
		return "", err
	}

	targetPath := targetPathRaw.ToString()

	return targetPath, nil
}

// createLink tries to create a shortcut in the given path that points to the given executable.
//
// Returns an error if it fails to create the shortcut.
func createLink(exePath string, linkPath string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	if err != nil {
		return err
	}
	defer ole.CoUninitialize()

	oleShellObject, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return err
	}
	defer oleShellObject.Release()

	wshell, err := oleShellObject.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer wshell.Release()

	cs, err := oleutil.CallMethod(wshell, "CreateShortcut", linkPath)
	if err != nil {
		return err
	}
	defer cs.Clear()

	idispatch := cs.ToIDispatch()

	oleutil.PutProperty(idispatch, "TargetPath", exePath)
	oleutil.CallMethod(idispatch, "Save")

	return nil
}

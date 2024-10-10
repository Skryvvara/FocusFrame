package main

import (
	"fmt"
	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"golang.org/x/sys/windows"
	"syscall"
	"time"
	"unsafe"
)

var Version string = "vX.Y.Z" // this is set during build time

var (
	enumWindows         = user32.NewProc("EnumWindows")
	user32              = windows.NewLazySystemDLL("user32.dll")
	findWindow          = user32.NewProc("FindWindowW")
	setWindowPos        = user32.NewProc("SetWindowPos")
	setWindowLong       = user32.NewProc("SetWindowLongW")
	getWindowLong       = user32.NewProc("GetWindowLongW")
	getForegroundWindow = user32.NewProc("GetForegroundWindow")
	getWindowText       = user32.NewProc("GetWindowTextW")
	getWindowTextLength = user32.NewProc("GetWindowTextLengthW")
	getAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
)

const (
	GWL_STYLE      = 0xFFFFFFFFFFFFFFF0
	WS_POPUP       = 0x00000000
	WS_VISIBLE     = 0x10000000
	WS_CAPTION     = 0x00C00000 // Title bar
	WS_THICKFRAME  = 0x00040000 // Resizable border
	SWP_NOZORDER   = 0x0004
	SWP_NOACTIVATE = 0x0010
)

var toggled bool = false

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("FocusFrame")
	systray.SetTooltip(fmt.Sprintf("FocusFrame Version: %s", Version))

	mAddGameCh := systray.AddMenuItem("Add Game", "Add Game to App")

	systray.AddSeparator()

	mRunOnStartupCh := systray.AddMenuItem("Run on Startup", "Run the application when starting the PC")

	systray.AddSeparator()

	mQuitCh := systray.AddMenuItem("Quit", "Quit the whole app")

	for {
		select {
		case <-mAddGameCh.ClickedCh:
			addGame()
		case <-mRunOnStartupCh.ClickedCh:
			if mRunOnStartupCh.Checked() {
				mRunOnStartupCh.Uncheck()
			} else {
				mRunOnStartupCh.Check()
			}
		case <-mQuitCh.ClickedCh:
			fmt.Println("Requesting Exit")
			systray.Quit()
		}
	}
}

func onExit() {
	// clean up here
}

func addGame() {
	//enumWindows.Call(syscall.NewCallback(enumWindowsProc), 0)

	title := selectWin()

	// Change "Window Title" to the title of the window you want to display

	hWnd, _, err := findWindow.Call(0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))))
	if err != nil && hWnd == 0 {
		fmt.Println("Window not found.")
		return
	}

	// Get the current window style
	currentStyle, _, _ := getWindowLong.Call(hWnd, uintptr(GWL_STYLE))

	// Remove the title bar and resizable border
	newStyle := (currentStyle &^ (WS_CAPTION | WS_THICKFRAME)) | WS_POPUP | WS_VISIBLE

	// Apply the new window style
	_, _, _ = setWindowLong.Call(hWnd, uintptr(GWL_STYLE), newStyle)

	// Set your desired fake fullscreen dimensions
	width := 2560
	height := 1440

	// Set the window position (x, y) and size (width, height)
	result, _, err := setWindowPos.Call(hWnd, 0, 1280, 0, uintptr(width), uintptr(height), SWP_NOZORDER|SWP_NOACTIVATE)
	if result == 0 {
		fmt.Printf("Failed to set window position. Error: %v\n", err)
	} else {
		fmt.Println("Window position set successfully.")
	}
}

func enumWindowsProc(hWnd uintptr, lParam uintptr) uintptr {
	length, _, _ := getWindowTextLength.Call(hWnd)
	if length == 0 {
		return 1
	}

	buf := make([]uint16, length+1)
	getWindowText.Call(hWnd, uintptr(unsafe.Pointer(&buf[0])), length+1)
	title := syscall.UTF16ToString(buf)

	fmt.Println(title)
	return 1
}

func selectWin() string {
	for {
		// Check if F3 key (virtual key code 114) is pressed
		keyState, _, _ := getAsyncKeyState.Call(114)
		if keyState&0x8000 != 0 {
			hWnd, _, _ := getForegroundWindow.Call()
			if hWnd == 0 {
				fmt.Println("No active window.")
				continue
			}

			length, _, _ := getWindowTextLength.Call(hWnd)
			if length == 0 {
				fmt.Println("No title for the active window.")
				continue
			}

			buf := make([]uint16, length+1)
			getWindowText.Call(hWnd, uintptr(unsafe.Pointer(&buf[0])), length+1)

			title := syscall.UTF16ToString(buf)
			fmt.Printf("Active window title: %s\n", title)

			return title
		}

		// Sleep briefly to avoid high CPU usage
		time.Sleep(100 * time.Millisecond)
	}
}

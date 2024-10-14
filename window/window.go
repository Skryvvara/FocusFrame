package window

import (
	"fmt"
	"github.com/StackExchange/wmi"
	"github.com/skryvvara/focusframe/config"
	"github.com/skryvvara/focusframe/input"
	"github.com/skryvvara/focusframe/process"
	"golang.org/x/sys/windows"
	"log"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32 = windows.NewLazySystemDLL("user32.dll")

	procFindWindow               = user32.NewProc("FindWindowW")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
	procSetWindowLong            = user32.NewProc("SetWindowLongW")
	procGetWindowLong            = user32.NewProc("GetWindowLongW")
	procGetWindowRect            = user32.NewProc("GetWindowRect")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowTextLength      = user32.NewProc("GetWindowTextLengthW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
)

const (
	GWL_EXSTYLE      = 0xFFFFFFFFFFFFFFEC // Offset for extended window styles
	GWL_STYLE        = 0xFFFFFFFFFFFFFFF0 // Style for tool windows
	WS_POPUP         = 0x00000000
	WS_EX_TOOLWINDOW = 0x00000080
	WS_VISIBLE       = 0x10000000
	WS_CAPTION       = 0x00C00000 // Title bar
	WS_THICKFRAME    = 0x00040000 // Resizable border
	SWP_NOZORDER     = 0x0004
	SWP_NOACTIVATE   = 0x0010
)

type RECT struct {
	Left, Top, Right, Bottom int32
}

type WindowInfo struct {
	Title string
	Hwnd  syscall.Handle
}

var OpenWindows []WindowInfo

// GetWindowText retrieves the text of the window identified by the handle
func GetWindowText(hwnd uintptr) string {
	buf := make([]uint16, 256)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return windows.UTF16ToString(buf)
}

func GetWindowByProcessID(pid uint32) syscall.Handle {
	OpenWindows = nil // Clear previous windows

	procEnumWindows.Call(syscall.NewCallback(EnumWindowsCallback), 0)

	for _, win := range OpenWindows {
		var windowPid uint32
		procGetWindowThreadProcessId.Call(uintptr(win.Hwnd), uintptr(unsafe.Pointer(&windowPid)))

		if windowPid == pid {
			return win.Hwnd
		}
	}
	return 0
}

// GetForegroundWindow gets the handle to the foreground window
func GetForegroundWindow() uintptr {
	handle, _, _ := procGetForegroundWindow.Call()
	return handle
}

func printWindowList() {
	procEnumWindows.Call(syscall.NewCallback(EnumWindowsCallback), 0)

	// Example: Find the executable for "Windows PowerShell" window
	for _, win := range OpenWindows {
		fmt.Printf("Found PowerShell Window: Handle %v\n", win.Hwnd)
		pid, err := process.GetProcessIDFromWindow(win.Hwnd)
		if err != nil {
			fmt.Println("Error getting PID:", err)
			continue
		}

		executable, err := process.GetExecutableFromPID(pid)
		if err != nil {
			fmt.Println("Error getting executable:", err)
			continue
		}

		fmt.Printf("%s Executable: %s\n", win.Title, executable)
	}
}

// EnumWindowsCallback is the callback function for EnumWindows
func EnumWindowsCallback(hwnd syscall.Handle, lParam uintptr) uintptr {
	// Check if the window is visible
	visible, _, err := procIsWindowVisible.Call(uintptr(hwnd))
	if err != syscall.Errno(0) || visible == 0 {
		return 1 // Skip if the window is not visible
	}

	// Filter out tool windows or other unwanted window styles
	if isToolWindow(hwnd) {
		return 1 // Skip if the window is a tool window
	}

	// Get the length of the window title
	length, _, _ := procGetWindowTextLength.Call(uintptr(hwnd))
	if length == 0 {
		return 1
	}

	// Allocate a buffer for the window title
	title := make([]uint16, length+1)

	// Get the window title
	_, _, _ = procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&title[0])), uintptr(length+1))

	// Convert the title to a string
	windowTitle := syscall.UTF16ToString(title)

	// Ignore windows with empty or unwanted titles
	if windowTitle == "" {
		return 1
	}

	// Add to the list of windows if the title is valid
	OpenWindows = append(OpenWindows, WindowInfo{Title: windowTitle, Hwnd: hwnd})

	return 1
}

// IsToolWindow checks if the window has the WS_EX_TOOLWINDOW style (tool windows)
func isToolWindow(hwnd syscall.Handle) bool {
	exStyle, _, _ := procGetWindowLong.Call(uintptr(hwnd), uintptr(GWL_EXSTYLE))

	// Check if the WS_EX_TOOLWINDOW bit is set
	if exStyle&WS_EX_TOOLWINDOW != 0 {
		return true
	}
	return false
}

func WatchForegroundWindowChange() {
	var lastWindow uintptr
	for {
		// Get the handle of the current foreground window
		currentWindow := GetForegroundWindow()

		// If the window handle has changed, we trigger the event
		if currentWindow != lastWindow && currentWindow != 0 {
			executable, err := process.GetExecutableFromHandle(syscall.Handle(currentWindow))
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Foreground window changed! New window exe: %s\n", executable)

			for _, game := range config.Config.ManagedApps {
				if game.Executable == executable {
					MoveWindow(executable)
				}
			}

			// Update the last window handle
			lastWindow = currentWindow
		}

		// Sleep for a short duration before checking again
		time.Sleep(500 * time.Millisecond)
	}
}

// Deprecated: This function is deprecated and may be removed in future versions.
// Use WatchForegroundWindowChange instead for improved performance and reliability.
//
// WatchProcessStart scans all open (and visible) windows every five seconds and moves
// them if they are managed. While this provides some functionality, it has a severe
// performance impact. It is advised to use WatchForegroundWindowChange instead,
// which is more efficient. Consider transitioning to this new implementation.
func WatchProcessStart() {
	for {
		var processes []process.Win32_Process
		query := "SELECT ProcessID, Name, ExecutablePath FROM Win32_Process"
		err := wmi.Query(query, &processes)
		if err != nil {
			fmt.Println("Error querying WMI:", err)
			continue
		}

		for _, proc := range processes {

			for _, app := range config.Config.ManagedApps {
				if strings.ToLower(app.Executable) == strings.ToLower(proc.Name) {
					fmt.Printf("Process started: %s, PID: %d\n", proc.Name, proc.ProcessID)

					// Now we call addGame to modify the window based on the process
					MoveWindow(proc.Name)
				}
			}
		}

		time.Sleep(5 * time.Second) // Poll every 5 seconds
	}
}

// Deprecated: This function is deprecated and may be removed in future versions.
// Use AddAppOnKeyPress instead for better user experience.
//
// selectWin waits for the user to press F3 while having the desired window focused
// to call moveWindow on it. However, this implementation is unintuitive for the end user.
// Consider transitioning to AddAppOnKeyPress for improved usability.
// This function might be removed in a future release.
func selectWin() string {
	for {
		// Check if F3 key (virtual key code 114) is pressed
		if input.IsKeyPressed(input.VK_F3) {
			hWnd, _, _ := procGetForegroundWindow.Call()
			if hWnd == 0 {
				fmt.Println("No active window.")
				continue
			}

			length, _, _ := procGetWindowTextLength.Call(hWnd)
			if length == 0 {
				fmt.Println("No title for the active window.")
				continue
			}

			buf := make([]uint16, length+1)
			procGetWindowTextW.Call(hWnd, uintptr(unsafe.Pointer(&buf[0])), length+1)

			title := syscall.UTF16ToString(buf)
			fmt.Printf("Active window title: %s\n", title)

			return title
		}

		// Sleep briefly to avoid high CPU usage
		time.Sleep(100 * time.Millisecond)
	}
}

func AddAppOnKeyPress(keyCode int) {
	for {
		// Check if the F4 key is pressed
		if input.IsKeyPressed(keyCode) { // If the most significant bit is set, the key is pressed
			fmt.Println(config.Config.ManagedApps)
			currentWindow := GetForegroundWindow()

			executable, err := process.GetExecutableFromHandle(syscall.Handle(currentWindow))
			if err != nil {
				log.Fatal(err)
			}

			hit := false
			for _, app := range config.Config.ManagedApps {
				if app.Executable == executable {
					hit = true
				}
			}

			if !hit {
				config.AddApplication(executable)
				MoveWindow(executable)
			} else {
				config.RemoveApplication(executable)
			}
		}

		// Sleep for a short duration to avoid excessive CPU usage
		time.Sleep(100 * time.Millisecond)
	}
}

// MoveWindow tries to find a window handle for the given executable and sets the window style and
// dimensions for the window.
//
// If the style and dimensions are already set, nothing is done.
func MoveWindow(executable string) {
	pid := process.GetProcessIDByExecutable(executable) // Find the process ID by the executable
	if pid == 0 {
		fmt.Println("Process not found.")
		return
	}

	hWnd := GetWindowByProcessID(pid) // Find the window handle by process ID
	if hWnd == 0 {
		fmt.Println("Window not found.")
		return
	}

	// Get the current window style
	currentStyle, _, _ := procGetWindowLong.Call(uintptr(hWnd), uintptr(GWL_STYLE))

	// Desired window style (remove title bar and resizable border)
	desiredStyle := (currentStyle &^ (WS_CAPTION | WS_THICKFRAME)) | WS_POPUP | WS_VISIBLE

	// Only apply the new window style if it's different from the current one
	if currentStyle != desiredStyle {
		_, _, _ = procSetWindowLong.Call(uintptr(hWnd), uintptr(GWL_STYLE), desiredStyle)
		fmt.Println("Window style updated.")
	} else {
		fmt.Println("Window style already correct, no changes needed.")
	}

	ws := config.GetWindowSettings(executable)

	// Get the current window rectangle (position and size)
	var rect RECT
	_, _, err := procGetWindowRect.Call(uintptr(hWnd), uintptr(unsafe.Pointer(&rect)))
	if err != syscall.Errno(0) {
		fmt.Printf("Failed to get window rectangle: %v\n", err)
		return
	}

	currentWidth := int(rect.Right - rect.Left)
	currentHeight := int(rect.Bottom - rect.Top)
	currentX := int(rect.Left)
	currentY := int(rect.Top)

	// Only set the window position and size if they are different from the desired values
	if currentWidth != ws.Width || currentHeight != ws.Height || currentX != ws.OffsetX || currentY != ws.OffsetY {
		result, _, err := procSetWindowPos.Call(uintptr(hWnd), 0, uintptr(ws.OffsetX), uintptr(ws.OffsetY), uintptr(ws.Width), uintptr(ws.Height), SWP_NOZORDER|SWP_NOACTIVATE)
		if result == 0 {
			fmt.Printf("Failed to set window position. Error: %v\n", err)
		} else {
			fmt.Println("Window position and size updated.")
		}
	} else {
		fmt.Println("Window position and size already correct, no changes needed.")
	}
}

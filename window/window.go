package window

import (
	"log"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/StackExchange/wmi"
	"github.com/lxn/win"
	"github.com/skryvvara/focusframe/config"
	"github.com/skryvvara/focusframe/input"
	"github.com/skryvvara/focusframe/process"
	"golang.org/x/sys/windows"
)

var (
	user32 = windows.NewLazySystemDLL("user32.dll")

	procFindWindow               = user32.NewProc("FindWindowW")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
	procSetWindowLongW           = user32.NewProc("SetWindowLongW")
	procGetWindowLongW           = user32.NewProc("GetWindowLongW")
	procGetWindowRect            = user32.NewProc("GetWindowRect")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
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

	WINEVENT_OUTOFCONTEXT   = 0x0000
	EVENT_SYSTEM_FOREGROUND = 0x0003
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
//
// This function uses the GetWindowTextW function from winuser.h.
//
// See https://learn.microsoft.com/de-de/windows/win32/api/winuser/nf-winuser-getwindowtextw
func GetWindowText(hwnd uintptr) string {
	buf := make([]uint16, 256)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return windows.UTF16ToString(buf)
}

// getWindowByProcessID tries to get the window beloging to the given PID.
// On success the handle of the window is returned otherwise the return value is 0.
//
// This function uses the GetWindowThreadProcessId function from winuser.h.
//
// See https://learn.microsoft.com/de-de/windows/win32/api/winuser/nf-winuser-getwindowthreadprocessid
func GetWindowByProcessID(pid uint32) syscall.Handle {
	OpenWindows = nil

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
//
// This function uses the GetForegroundWindow function from winuser.h.
//
// See https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getforegroundwindow
func GetForegroundWindow() uintptr {
	handle, _, _ := procGetForegroundWindow.Call()
	return handle
}

// Deprecated: This function was a PoC to print a list of currently open
// windows and is currently unneeded and unused but kept for potential use
// later on for the v1.0.0 Version featuring a GUI to select the applications
// to be managed via a list.
func printWindowList() {
	procEnumWindows.Call(syscall.NewCallback(EnumWindowsCallback), 0)

	for _, win := range OpenWindows {
		pid, err := process.GetProcessIDFromWindow(win.Hwnd)
		if err != nil {
			log.Println("Error getting PID:", err)
			continue
		}

		executable, err := process.GetExecutableFromPID(pid)
		if err != nil {
			log.Println("Error getting executable:", err)
			continue
		}

		log.Printf("%s Executable: %s\n", win.Title, executable)
	}
}

// EnumWindowsCallback is the callback function for EnumWindows
//
// This function uses the GetWindowTextW and GetWindowTextLengthW functions from winuser.h.
//
// See https://learn.microsoft.com/de-de/windows/win32/api/winuser/nf-winuser-getwindowtextw
// and https://learn.microsoft.com/de-de/windows/win32/api/winuser/nf-winuser-getwindowtextlengthw
func EnumWindowsCallback(hwnd syscall.Handle, lParam uintptr) uintptr {
	visible, _, err := procIsWindowVisible.Call(uintptr(hwnd))
	if err != syscall.Errno(0) || visible == 0 {
		return 1
	}

	if isToolWindow(hwnd) {
		return 1
	}

	length, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
	if length == 0 {
		return 1
	}

	title := make([]uint16, length+1)

	_, _, _ = procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&title[0])), uintptr(length+1))

	windowTitle := syscall.UTF16ToString(title)

	if windowTitle == "" {
		return 1
	}

	OpenWindows = append(OpenWindows, WindowInfo{Title: windowTitle, Hwnd: hwnd})

	return 1
}

// IsToolWindow checks if the window has the WS_EX_TOOLWINDOW style (tool windows)
//
// See https://learn.microsoft.com/en-us/windows/win32/winmsg/extended-window-styles
func isToolWindow(hwnd syscall.Handle) bool {
	exStyle, _, _ := procGetWindowLongW.Call(uintptr(hwnd), uintptr(GWL_EXSTYLE))

	return exStyle&WS_EX_TOOLWINDOW != 0

}

// ForegroundWindowEvent is called when the foreground window changes
func ForegroundWindowEvent(hWinEventHook win.HWINEVENTHOOK, event uint32, hwnd win.HWND, idObject int32, idChild int32, idEventThread uint32, dwmsEventTime uint32) uintptr {
	executable, err := process.GetExecutableFromHandle(syscall.Handle(uintptr(hwnd)))
	if err != nil {
		log.Println("Error getting executable:", err)
		return 1
	}

	log.Printf("Foreground window changed! New window exe: %s\n", executable)

	for _, game := range config.Config.ManagedApps {
		if game.Executable == executable {
			MoveWindow(executable)
		}
	}

	return 0
}

// CreateWinEventHook sets up the SetWinEventHook for foreground window changes
func CreateWinEventHook() win.HWINEVENTHOOK {
	cb := win.WINEVENTPROC(ForegroundWindowEvent)

	hook, err := win.SetWinEventHook(
		EVENT_SYSTEM_FOREGROUND,
		EVENT_SYSTEM_FOREGROUND,
		0,
		cb,
		0,
		0,
		WINEVENT_OUTOFCONTEXT,
	)
	if err != nil {
		log.Println("Error set window event hook:", err)
		return 0
	}
	if hook == 0 {
		log.Println("Failed to set hook.")
	}
	return hook
}

// WatchForegroundWindowChange starts listening for foreground window change events
func WatchForegroundWindowChange() {
	log.Println("Starting to watch foreground window changes")
	hook := CreateWinEventHook()
	if hook == 0 {
		log.Println("Failed to create foreground window hook.")
		return
	}
	defer win.UnhookWinEvent(hook)

	// Run a basic Windows message loop to keep the program listening
	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) > 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
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
			log.Println("Error querying WMI:", err)
			continue
		}

		for _, proc := range processes {

			for _, app := range config.Config.ManagedApps {
				if strings.EqualFold(app.Executable, proc.Name) {
					log.Printf("Process started: %s, PID: %d\n", proc.Name, proc.ProcessID)

					MoveWindow(proc.Name)
				}
			}
		}

		time.Sleep(5 * time.Second)
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
		if input.IsKeyPressed(input.VK_F3) {
			hWnd, _, _ := procGetForegroundWindow.Call()
			if hWnd == 0 {
				log.Println("No active window.")
				continue
			}

			length, _, _ := procGetWindowTextLengthW.Call(hWnd)
			if length == 0 {
				log.Println("No title for the active window.")
				continue
			}

			buf := make([]uint16, length+1)
			procGetWindowTextW.Call(hWnd, uintptr(unsafe.Pointer(&buf[0])), length+1)

			title := syscall.UTF16ToString(buf)
			log.Printf("Active window title: %s\n", title)

			return title
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// TODO: name is misleading since it also removes apps and the hotkey thing is not super user friendly.
// AddAppOnKeyPress takes the keyCode and waits for the key to be pressed.
// When pressed, the currently focused window is taken and either added to list of
// managed applications or removed from it if it is already on the list.
// Should this fail, the case will be ignored and the function waits on the next keypress.
func AddAppOnKeyPress(keyCode int) {
	for {
		if input.IsKeyPressed(keyCode) {
			log.Println(config.Config.ManagedApps)
			currentWindow := GetForegroundWindow()

			executable, err := process.GetExecutableFromHandle(syscall.Handle(currentWindow))
			if err != nil {
				log.Println(err)
				continue
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

		time.Sleep(100 * time.Millisecond)
	}
}

// getWindowStyle returns the window style as an uintptr.
//
// This function uses the GetWindowLongW function from winuser.h.
//
// See https://learn.microsoft.com/de-de/windows/win32/api/winuser/nf-winuser-getwindowlongw
func getWindowStyle(hWnd syscall.Handle) uintptr {
	style, _, _ := procGetWindowLongW.Call(uintptr(hWnd), uintptr(GWL_STYLE))
	return style
}

// setWindowStyle sets the (hardcoded) window style to the given handle.
//
// This function uses the SetWindowLongW function from winuser.h.
//
// See https://learn.microsoft.com/de-de/windows/win32/api/winuser/nf-winuser-setwindowlongw
func setWindowStyle(hWnd syscall.Handle) {
	currentStyle := getWindowStyle(hWnd)

	desiredStyle := (currentStyle &^ (WS_CAPTION | WS_THICKFRAME)) | WS_POPUP | WS_VISIBLE

	if currentStyle != desiredStyle {
		_, _, _ = procSetWindowLongW.Call(uintptr(hWnd), uintptr(GWL_STYLE), desiredStyle)
		log.Println("Window style updated.")
	} else {
		log.Println("Window style already correct, no changes needed.")
	}
}

// getWindowRect tries to get the rect of the window with the given handle.
// On success the rect is returned and the error is nil.
//
// This function uses the GetWindowRect function from winuser.h.
//
// See https://learn.microsoft.com/de-de/windows/win32/api/winuser/nf-winuser-getwindowrect
func getWindowRect(hWnd syscall.Handle) (RECT, error) {
	var rect RECT
	_, _, err := procGetWindowRect.Call(uintptr(hWnd), uintptr(unsafe.Pointer(&rect)))
	if err != syscall.Errno(0) {
		return rect, err
	}
	return rect, nil
}

// setWindowRect tries to set the window position and size.
// On failure an error is returned otherwise the return value is nil.
func setWindowPos(hWnd syscall.Handle, ws config.WindowSettings) error {
	rect, err := getWindowRect(hWnd)
	if err != nil {
		return err
	}

	if int(rect.Right-rect.Left) == ws.Width &&
		int(rect.Bottom-rect.Top) == ws.Height &&
		int(rect.Left) == ws.OffsetX &&
		int(rect.Top) == ws.OffsetY {
		log.Println("Window position and size already correct, no changes needed.")
		return nil
	}

	result, err := callSetWindowPosProc(hWnd, ws)
	if result == 0 {
		return err
	}

	// This fixes #47, I don't have a better fix currently but this will do for now
	for i := 0; i < 3; i++ {
		result, err := callSetWindowPosProc(hWnd, ws)
		if result == 0 {
			log.Println(err)
			time.Sleep(100 * time.Millisecond) // Short delay between retries
			continue
		}
		break
	}

	return nil
}

// callSetWindowPosProc is a wrapper for the call to procSetWindowPos to reduce redunancy in
// setWindowPos. On success the result is returned otherwise the result (value 0) and an error
// is returned.
//
// This function uses the SetWindow function from winuser.h.
//
// See https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setwindowpos
func callSetWindowPosProc(hWnd syscall.Handle, ws config.WindowSettings) (uintptr, error) {
	result, _, err := procSetWindowPos.Call(
		uintptr(hWnd),
		0,
		uintptr(ws.OffsetX),
		uintptr(ws.OffsetY),
		uintptr(ws.Width),
		uintptr(ws.Height),
		uintptr(SWP_NOZORDER|SWP_NOACTIVATE),
	)

	if result == 0 {
		return result, err
	}
	return result, nil
}

// MoveWindow tries to find a window handle for the given executable and sets the window style and
// dimensions for the window.
//
// If the style and dimensions are already set, nothing is done.
func MoveWindow(executable string) {
	pid := process.GetProcessIDByExecutable(executable) // Find the process ID by the executable
	if pid == 0 {
		log.Println("Process not found.")
		return
	}

	hWnd := GetWindowByProcessID(pid) // Find the window handle by process ID
	if hWnd == 0 {
		log.Println("Window not found.")
		return
	}

	setWindowStyle(hWnd)

	ws := config.GetWindowSettings(executable)

	err := setWindowPos(hWnd, ws)
	if err != nil {
		log.Println(err)
		return
	}
}

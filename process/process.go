package process

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows"
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	psapi    = syscall.NewLazyDLL("psapi.dll")

	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procOpenProcess              = kernel32.NewProc("OpenProcess")
	procGetModuleFileNameEx      = psapi.NewProc("GetModuleFileNameExW")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
)

const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
)

// Win32_Process WMI class structure
type Win32_Process struct {
	ProcessID uint32
	Name      string
}

// GetProcessIDByExecutable tries to find the PID (Process ID) of a process with
// the given executable.
//
// Returns either the PID as a uint32 or an error if no PID could be found.
func GetProcessIDByExecutable(executable string) (uint32, error) {
	var processes []Win32_Process
	query := fmt.Sprintf("SELECT ProcessID, Name FROM Win32_Process WHERE Name='%s'", executable)
	err := wmi.Query(query, &processes)
	if err != nil {
		return 0, fmt.Errorf("error querying WMI: %v", err)
	}

	if len(processes) > 0 {
		return processes[0].ProcessID, nil
	}
	return 0, nil
}

// GetProcessIDFromWindow tries to find the PID (Process ID) of a process with
// the given window handle.
//
// Returns either the PID as a uint32 or an error if no PID could be found.
func GetProcessIDFromWindow(hWnd uintptr) (uint32, error) {
	var pid uint32
	_, _, err := procGetWindowThreadProcessId.Call(hWnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return 0, fmt.Errorf("failed to get PID: %v", err)
	}
	return pid, nil
}

// GetFullExecutableFromPID tries to get the full path of the executable of a process
// with the given PID (Process ID). E.g. C:/Applications/Program.exe
//
// Returns either the Executable as a string or an error if the executable could not be found.
func GetFullExecutableFromPID(pid uint32) (string, error) {
	hProcess, _, _ := procOpenProcess.Call(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ, 0, uintptr(pid))
	if hProcess == 0 {
		return "", fmt.Errorf("failed to open process: %v", syscall.GetLastError())
	}
	defer func() {
		ret, _, err := procCloseHandle.Call(hProcess)
		if ret == 0 {
			fmt.Printf("Failed to close handle: %v\n", err)
		}
	}()

	exePath := make([]uint16, syscall.MAX_PATH)
	_, _, err := procGetModuleFileNameEx.Call(hProcess, 0, uintptr(unsafe.Pointer(&exePath[0])), syscall.MAX_PATH)
	if err != syscall.Errno(0) {
		return "", fmt.Errorf("failed to get module file name: %v", err)
	}

	return syscall.UTF16ToString(exePath), nil
}

// GetExecutableFromPID functions like GetFullExectuableFromPID but returns only the executable's name.
// E.g Program.exe
//
// For more information, see GetFullExecutableFromPID.
//
// Returns either the Executable as a string or an error if the executable could not be found.
func GetExecutableFromPID(pid uint32) (string, error) {
	fullPath, err := GetFullExecutableFromPID(pid)
	if err != nil {
		return "", err
	}

	// Extract the base name from the full path (e.g., "bg3.exe" from "F:\path\to\bg3.exe")
	baseName := filepath.Base(fullPath)

	return baseName, nil
}

// GetExecutableFromHandle tries to get the (short) executable name of a window
// given the window handle.
//
// Returns either the Executable as a string or an error if the executable could not be found.
func GetExecutableFromHandle(hWnd uintptr) (string, error) {
	pid, err := GetProcessIDFromWindow(hWnd)
	if err != nil {
		return "", err
	}

	executable, err := GetExecutableFromPID(pid)
	if err != nil {
		return "", err
	}

	return executable, nil
}

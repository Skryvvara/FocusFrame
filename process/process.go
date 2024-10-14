package process

import (
	"fmt"
	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows"
	"path/filepath"
	"syscall"
	"unsafe"
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

func GetProcessIDByExecutable(executable string) uint32 {
	var processes []Win32_Process
	query := fmt.Sprintf("SELECT ProcessID, Name FROM Win32_Process WHERE Name='%s'", executable)
	err := wmi.Query(query, &processes)
	if err != nil {
		fmt.Println("Error querying WMI:", err)
		return 0
	}

	if len(processes) > 0 {
		return processes[0].ProcessID
	}
	return 0
}

func GetProcessIDFromWindow(hwnd syscall.Handle) (uint32, error) {
	var pid uint32
	_, _, err := procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return 0, fmt.Errorf("failed to get PID: %v", err)
	}
	return pid, nil
}

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

func GetExecutableFromPID(pid uint32) (string, error) {
	hProcess, _, _ := procOpenProcess.Call(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ, 0, uintptr(pid))
	if hProcess == 0 {
		return "", fmt.Errorf("failed to open process: %v", syscall.GetLastError())
	}
	defer procCloseHandle.Call(hProcess)

	exePath := make([]uint16, syscall.MAX_PATH)
	_, _, err := procGetModuleFileNameEx.Call(hProcess, 0, uintptr(unsafe.Pointer(&exePath[0])), syscall.MAX_PATH)
	if err != syscall.Errno(0) {
		return "", fmt.Errorf("failed to get module file name: %v", err)
	}

	fullPath := syscall.UTF16ToString(exePath)
	// Extract the base name from the full path (e.g., "bg3.exe" from "F:\path\to\bg3.exe")
	baseName := filepath.Base(fullPath)

	return baseName, nil
}

func GetExecutableFromHandle(hwnd syscall.Handle) (string, error) {
	pid, err := GetProcessIDFromWindow(hwnd)
	if err != nil {
		return "", err
	}

	executable, err := GetExecutableFromPID(pid)
	if err != nil {
		return "", err
	}

	return executable, nil
}

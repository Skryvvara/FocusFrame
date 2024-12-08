package input

import (
	"syscall"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")
)

// Define virtual key codes for the keys you are interested in.
const (
	VK_F3        = 0x72 // F3 key virtual key code
	VK_F4        = 0x73 // F4 key virtual key code
	VK_NUM_SLASH = 0x6F // Numpad Slash virtual key code
	// Add other keys as needed
)

// IsKeyPressed checks if a key is currently pressed.
// Returns true if the key is pressed, false otherwise.
func IsKeyPressed(vkCode int) bool {
	keyState, _, _ := procGetAsyncKeyState.Call(uintptr(vkCode))
	return keyState&0x8000 != 0
}

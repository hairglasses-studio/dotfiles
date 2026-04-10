//go:build windows

package output

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/hairglasses-studio/mapitall/internal/mapping"
)

var (
	user32    = syscall.NewLazyDLL("user32.dll")
	sendInput = user32.NewProc("SendInput")
)

const (
	inputKeyboard      = 1
	keybdEventKeyUp    = 0x0002
	keybdEventScanCode = 0x0008
)

type keyboardInput struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

type input struct {
	inputType uint32
	ki        keyboardInput
	_padding  [8]byte
}

// KeyTarget emits virtual key events via Windows SendInput.
type KeyTarget struct{}

func NewKeyTarget() *KeyTarget { return &KeyTarget{} }

func (t *KeyTarget) Type() mapping.OutputType { return mapping.OutputKey }

func (t *KeyTarget) Execute(action mapping.OutputAction, value float64) error {
	pressed := value > 0.5

	for _, keyName := range action.Keys {
		vk, ok := WindowsKeyCodes[keyName]
		if !ok {
			return fmt.Errorf("unknown Windows key: %s", keyName)
		}

		var flags uint32
		if !pressed {
			flags = keybdEventKeyUp
		}

		inp := input{
			inputType: inputKeyboard,
			ki: keyboardInput{
				wVk:     vk,
				dwFlags: flags,
			},
		}

		ret, _, _ := sendInput.Call(
			1,
			uintptr(unsafe.Pointer(&inp)),
			unsafe.Sizeof(inp),
		)
		if ret == 0 {
			return fmt.Errorf("SendInput failed for key %s", keyName)
		}
	}
	return nil
}

func (t *KeyTarget) Close() error { return nil }

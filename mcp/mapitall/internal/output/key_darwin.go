//go:build darwin

package output

/*
#cgo LDFLAGS: -framework CoreGraphics -framework Carbon
#include <CoreGraphics/CoreGraphics.h>

void postKeyEvent(int keyCode, int keyDown) {
    CGEventRef event = CGEventCreateKeyboardEvent(NULL, (CGKeyCode)keyCode, keyDown);
    CGEventPost(kCGHIDEventTap, event);
    CFRelease(event);
}
*/
import "C"

import (
	"fmt"

	"github.com/hairglasses-studio/mapitall/internal/mapping"
)

// KeyTarget emits virtual key events via macOS CoreGraphics.
type KeyTarget struct{}

func NewKeyTarget() *KeyTarget { return &KeyTarget{} }

func (t *KeyTarget) Type() mapping.OutputType { return mapping.OutputKey }

func (t *KeyTarget) Execute(action mapping.OutputAction, value float64) error {
	pressed := value > 0.5

	for _, keyName := range action.Keys {
		code, ok := DarwinKeyCodes[keyName]
		if !ok {
			return fmt.Errorf("unknown macOS key: %s", keyName)
		}

		if pressed {
			C.postKeyEvent(C.int(code), 1)
		} else {
			C.postKeyEvent(C.int(code), 0)
		}
	}
	return nil
}

func (t *KeyTarget) Close() error { return nil }

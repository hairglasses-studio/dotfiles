//go:build windows

package output

import (
	"context"
	"math"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/hairglasses-studio/mapping"
)

const (
	inputMouse     = 0
	mouseMove      = 0x0001
	mouseWheel     = 0x0800
	mouseHWheel    = 0x1000
	wheelDelta     = 120
	winDeadzone    = 0.05
	winSensitivity = 10.0
	winEmitHz      = 200
)

type mouseInput struct {
	dx          int32
	dy          int32
	mouseData   uint32
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

type mouseInputWrapper struct {
	inputType uint32
	mi        mouseInput
	_padding  [8]byte
}

// MovementTarget emits cursor and scroll events via Windows SendInput.
type MovementTarget struct {
	mu     sync.Mutex
	cursor struct{ x, y float64 }
	scroll struct{ x, y float64 }
	cancel context.CancelFunc
}

func NewMovementTarget() *MovementTarget {
	t := &MovementTarget{}
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	go t.emitLoop(ctx)
	return t
}

func (t *MovementTarget) Type() mapping.OutputType { return mapping.OutputMovement }

func (t *MovementTarget) Execute(action mapping.OutputAction, value float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch action.Target {
	case "CURSOR_UP":
		t.cursor.y = -value
	case "CURSOR_DOWN":
		t.cursor.y = value
	case "CURSOR_LEFT":
		t.cursor.x = -value
	case "CURSOR_RIGHT":
		t.cursor.x = value
	case "SCROLL_UP":
		t.scroll.y = value
	case "SCROLL_DOWN":
		t.scroll.y = -value
	case "SCROLL_LEFT":
		t.scroll.x = -value
	case "SCROLL_RIGHT":
		t.scroll.x = value
	}
	return nil
}

func (t *MovementTarget) emitLoop(ctx context.Context) {
	user32 := syscall.NewLazyDLL("user32.dll")
	sendInputProc := user32.NewProc("SendInput")

	ticker := time.NewTicker(time.Second / winEmitHz)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.mu.Lock()
			cx, cy := t.cursor.x, t.cursor.y
			sx, sy := t.scroll.x, t.scroll.y
			t.mu.Unlock()

			if math.Abs(cx) > winDeadzone || math.Abs(cy) > winDeadzone {
				dx := int32(cx * winSensitivity)
				dy := int32(cy * winSensitivity)
				if dx != 0 || dy != 0 {
					inp := mouseInputWrapper{
						inputType: inputMouse,
						mi: mouseInput{
							dx:      dx,
							dy:      dy,
							dwFlags: mouseMove,
						},
					}
					sendInputProc.Call(1, uintptr(unsafe.Pointer(&inp)), unsafe.Sizeof(inp))
				}
			}

			if math.Abs(sy) > winDeadzone {
				amount := int32(sy * float64(wheelDelta))
				inp := mouseInputWrapper{
					inputType: inputMouse,
					mi: mouseInput{
						mouseData: uint32(amount),
						dwFlags:   mouseWheel,
					},
				}
				sendInputProc.Call(1, uintptr(unsafe.Pointer(&inp)), unsafe.Sizeof(inp))
			}

			if math.Abs(sx) > winDeadzone {
				amount := int32(sx * float64(wheelDelta))
				inp := mouseInputWrapper{
					inputType: inputMouse,
					mi: mouseInput{
						mouseData: uint32(amount),
						dwFlags:   mouseHWheel,
					},
				}
				sendInputProc.Call(1, uintptr(unsafe.Pointer(&inp)), unsafe.Sizeof(inp))
			}
		}
	}
}

func (t *MovementTarget) Close() error {
	if t.cancel != nil {
		t.cancel()
	}
	return nil
}

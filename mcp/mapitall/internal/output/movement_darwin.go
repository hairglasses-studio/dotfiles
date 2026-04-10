//go:build darwin

package output

/*
#cgo LDFLAGS: -framework CoreGraphics
#include <CoreGraphics/CoreGraphics.h>

void postMouseMove(int dx, int dy) {
    CGEventRef event = CGEventCreate(NULL);
    CGPoint loc = CGEventGetLocation(event);
    CFRelease(event);

    CGPoint newLoc = CGPointMake(loc.x + dx, loc.y + dy);
    CGEventRef move = CGEventCreateMouseEvent(NULL, kCGEventMouseMoved, newLoc, 0);
    CGEventPost(kCGHIDEventTap, move);
    CFRelease(move);
}

void postScrollWheel(int dx, int dy) {
    CGEventRef scroll = CGEventCreateScrollWheelEvent(NULL, kCGScrollEventUnitPixel, 2, dy, dx);
    CGEventPost(kCGHIDEventTap, scroll);
    CFRelease(scroll);
}
*/
import "C"

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/hairglasses-studio/mapitall/internal/mapping"
)

// MovementTarget emits cursor movement and scroll events via macOS CoreGraphics.
type MovementTarget struct {
	mu     sync.Mutex
	cursor cursorPos
	scroll cursorPos
	cancel context.CancelFunc
}

type cursorPos struct {
	x, y float64
}

const (
	darwinDeadzone    = 0.05
	darwinSensitivity = 10.0
	darwinEmitHz      = 200
)

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
	ticker := time.NewTicker(time.Second / darwinEmitHz)
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

			if math.Abs(cx) > darwinDeadzone || math.Abs(cy) > darwinDeadzone {
				dx := int(cx * darwinSensitivity)
				dy := int(cy * darwinSensitivity)
				if dx != 0 || dy != 0 {
					C.postMouseMove(C.int(dx), C.int(dy))
				}
			}

			if math.Abs(sx) > darwinDeadzone || math.Abs(sy) > darwinDeadzone {
				dx := int(sx * darwinSensitivity)
				dy := int(sy * darwinSensitivity)
				if dx != 0 || dy != 0 {
					C.postScrollWheel(C.int(dx), C.int(dy))
				}
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

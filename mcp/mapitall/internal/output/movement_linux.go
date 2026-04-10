//go:build linux

package output

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/hairglasses-studio/mapitall/internal/mapping"
)

// Relative axis event codes (from linux/input-event-codes.h).
const (
	evRel         = 0x02
	relX          = 0x00
	relY          = 0x01
	relWheel      = 0x08
	relHWheel     = 0x06
	relWheelHiRes = 0x0b

	uiSetRelbit = 0x40045566 // UI_SET_RELBIT
)

// MovementTarget emits cursor/scroll events via Linux uinput.
type MovementTarget struct {
	fd     *os.File
	cancel context.CancelFunc

	mu        sync.Mutex
	positions map[string]*cursorPos // "cursor_x", "cursor_y", "scroll_x", "scroll_y"
}

type cursorPos struct {
	value       float64 // accumulated position
	sensitivity float64
	deadzone    float64
}

// NewMovementTarget creates a uinput relative axis device.
func NewMovementTarget() *MovementTarget {
	t := &MovementTarget{
		positions: make(map[string]*cursorPos),
	}
	if err := t.init(); err != nil {
		slog.Warn("uinput movement target init failed (movement output disabled)", "error", err)
	}
	return t
}

func (t *MovementTarget) init() error {
	f, err := os.OpenFile(uinputPath, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("open %s: %w", uinputPath, err)
	}

	// Enable EV_REL and EV_KEY (for BTN_LEFT etc).
	if err := ioctl(f, uiSetEvbit, evRel); err != nil {
		f.Close()
		return err
	}
	if err := ioctl(f, uiSetEvbit, evKey); err != nil {
		f.Close()
		return err
	}

	// Enable relative axes.
	for _, rel := range []uintptr{relX, relY, relWheel, relHWheel, relWheelHiRes} {
		if err := ioctl(f, uiSetRelbit, rel); err != nil {
			f.Close()
			return err
		}
	}

	// Enable mouse buttons for click simulation.
	for _, btn := range []uintptr{0x110, 0x111, 0x112} { // BTN_LEFT, BTN_RIGHT, BTN_MIDDLE
		if err := ioctl(f, uiSetKeybit, btn); err != nil {
			f.Close()
			return err
		}
	}

	setup := uinputSetup{
		ID: inputID{Bustype: 0x03, Vendor: 0x4847, Product: 0x0002, Version: 1},
	}
	copy(setup.Name[:], "mapitall-pointer")
	if err := ioctlPtr(f, uiDevSetup, unsafe.Pointer(&setup)); err != nil {
		f.Close()
		return err
	}

	if err := ioctl(f, uiDevCreate, 0); err != nil {
		f.Close()
		return err
	}

	time.Sleep(50 * time.Millisecond)

	t.fd = f

	// Start emission loop at 200Hz.
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	go t.emitLoop(ctx)

	slog.Info("uinput movement device created")
	return nil
}

func (t *MovementTarget) Type() mapping.OutputType { return mapping.OutputMovement }

func (t *MovementTarget) Execute(action mapping.OutputAction, value float64) error {
	if t.fd == nil {
		return fmt.Errorf("uinput movement device not initialized")
	}

	target := action.Target
	switch target {
	case "CURSOR_UP":
		t.setPosition("cursor_y", -value)
	case "CURSOR_DOWN":
		t.setPosition("cursor_y", value)
	case "CURSOR_LEFT":
		t.setPosition("cursor_x", -value)
	case "CURSOR_RIGHT":
		t.setPosition("cursor_x", value)
	case "SCROLL_UP":
		t.setPosition("scroll_y", -value)
	case "SCROLL_DOWN":
		t.setPosition("scroll_y", value)
	case "SCROLL_LEFT":
		t.setPosition("scroll_x", -value)
	case "SCROLL_RIGHT":
		t.setPosition("scroll_x", value)
	default:
		// Direct axis update: "cursor" sets both X/Y, "scroll" sets scroll.
		t.setPosition(target, value)
	}
	return nil
}

func (t *MovementTarget) setPosition(axis string, value float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	pos, ok := t.positions[axis]
	if !ok {
		pos = &cursorPos{sensitivity: 6, deadzone: 5}
		t.positions[axis] = pos
	}
	pos.value = value
}

// emitLoop runs at 200Hz and translates accumulated positions to REL events.
func (t *MovementTarget) emitLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.emitMovement()
		}
	}
}

func (t *MovementTarget) emitMovement() {
	t.mu.Lock()
	cx := t.getPos("cursor_x")
	cy := t.getPos("cursor_y")
	sx := t.getPos("scroll_x")
	sy := t.getPos("scroll_y")
	t.mu.Unlock()

	wrote := false

	if dx := applyDeadzoneAndSensitivity(cx); dx != 0 {
		t.writeEvent(evRel, relX, int32(dx))
		wrote = true
	}
	if dy := applyDeadzoneAndSensitivity(cy); dy != 0 {
		t.writeEvent(evRel, relY, int32(dy))
		wrote = true
	}
	if wx := applyDeadzoneAndSensitivity(sx); wx != 0 {
		t.writeEvent(evRel, relHWheel, int32(wx))
		wrote = true
	}
	if wy := applyDeadzoneAndSensitivity(sy); wy != 0 {
		t.writeEvent(evRel, relWheel, int32(-wy)) // Invert: positive = scroll down
		wrote = true
	}

	if wrote {
		t.writeEvent(evSyn, synReport, 0)
	}
}

func (t *MovementTarget) getPos(axis string) float64 {
	if pos, ok := t.positions[axis]; ok {
		return pos.value
	}
	return 0
}

func applyDeadzoneAndSensitivity(value float64) float64 {
	const deadzone = 0.05
	const sensitivity = 10.0

	if math.Abs(value) < deadzone {
		return 0
	}
	return value * sensitivity
}

func (t *MovementTarget) writeEvent(typ, code uint16, value int32) {
	if t.fd == nil {
		return
	}
	now := time.Now()
	ev := inputEvent{
		Sec:  now.Unix(),
		Usec: int64(now.Nanosecond() / 1000),
		Type: typ,
		Code: code,
		Val:  value,
	}
	buf := make([]byte, inputEventSize)
	writeInputEvent(buf, &ev)
	if _, err := t.fd.Write(buf); err != nil {
		slog.Warn("uinput write failed", "error", err)
	}
}

func writeInputEvent(buf []byte, ev *inputEvent) {
	_ = buf[23]
	buf[0] = byte(ev.Sec)
	buf[1] = byte(ev.Sec >> 8)
	buf[2] = byte(ev.Sec >> 16)
	buf[3] = byte(ev.Sec >> 24)
	buf[4] = byte(ev.Sec >> 32)
	buf[5] = byte(ev.Sec >> 40)
	buf[6] = byte(ev.Sec >> 48)
	buf[7] = byte(ev.Sec >> 56)
	buf[8] = byte(ev.Usec)
	buf[9] = byte(ev.Usec >> 8)
	buf[10] = byte(ev.Usec >> 16)
	buf[11] = byte(ev.Usec >> 24)
	buf[12] = byte(ev.Usec >> 32)
	buf[13] = byte(ev.Usec >> 40)
	buf[14] = byte(ev.Usec >> 48)
	buf[15] = byte(ev.Usec >> 56)
	buf[16] = byte(ev.Type)
	buf[17] = byte(ev.Type >> 8)
	buf[18] = byte(ev.Code)
	buf[19] = byte(ev.Code >> 8)
	buf[20] = byte(ev.Val)
	buf[21] = byte(ev.Val >> 8)
	buf[22] = byte(ev.Val >> 16)
	buf[23] = byte(ev.Val >> 24)
}

func (t *MovementTarget) Close() error {
	if t.cancel != nil {
		t.cancel()
	}
	if t.fd == nil {
		return nil
	}
	ioctl(t.fd, uiDevDestroy, 0)
	return t.fd.Close()
}

//go:build linux

package output

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"time"
	"unsafe"

	"github.com/hairglasses-studio/mapitall/internal/mapping"
)

// uinput ioctl constants (from linux/uinput.h).
const (
	uinputPath = "/dev/uinput"

	// Event types.
	evSyn = 0x00
	evKey = 0x01

	// Sync event codes.
	synReport = 0x00

	// ioctl commands.
	uiSetEvbit   = 0x40045564 // UI_SET_EVBIT
	uiSetKeybit  = 0x40045565 // UI_SET_KEYBIT
	uiDevCreate  = 0x5501     // UI_DEV_CREATE
	uiDevDestroy = 0x5502     // UI_DEV_DESTROY
	uiDevSetup   = 0x405C5503 // UI_DEV_SETUP
)

// uinputSetup matches struct uinput_setup from linux/uinput.h.
type uinputSetup struct {
	ID   inputID
	Name [80]byte
	_    uint32 // ff_effects_max
}

type inputID struct {
	Bustype uint16
	Vendor  uint16
	Product uint16
	Version uint16
}

// inputEvent matches struct input_event.
type inputEvent struct {
	Sec  int64
	Usec int64
	Type uint16
	Code uint16
	Val  int32
}

const inputEventSize = int(unsafe.Sizeof(inputEvent{}))

// KeyTarget emits keyboard key events via Linux uinput.
type KeyTarget struct {
	fd *os.File
}

// NewKeyTarget creates a uinput virtual keyboard device.
func NewKeyTarget() *KeyTarget {
	t := &KeyTarget{}
	if err := t.init(); err != nil {
		slog.Warn("uinput key target init failed (key output disabled)", "error", err)
	}
	return t
}

func (t *KeyTarget) init() error {
	f, err := os.OpenFile(uinputPath, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("open %s: %w (check udev rules / input group)", uinputPath, err)
	}

	// Enable EV_KEY.
	if err := ioctl(f, uiSetEvbit, evKey); err != nil {
		f.Close()
		return fmt.Errorf("UI_SET_EVBIT: %w", err)
	}

	// Enable all key codes we might emit.
	for _, code := range LinuxKeyCodes {
		if err := ioctl(f, uiSetKeybit, uintptr(code)); err != nil {
			f.Close()
			return fmt.Errorf("UI_SET_KEYBIT %d: %w", code, err)
		}
	}

	// Set up device metadata.
	setup := uinputSetup{
		ID: inputID{Bustype: 0x03, Vendor: 0x4847, Product: 0x0001, Version: 1}, // "HG" vendor
	}
	copy(setup.Name[:], "mapitall-keys")
	if err := ioctlPtr(f, uiDevSetup, unsafe.Pointer(&setup)); err != nil {
		f.Close()
		return fmt.Errorf("UI_DEV_SETUP: %w", err)
	}

	// Create the device.
	if err := ioctl(f, uiDevCreate, 0); err != nil {
		f.Close()
		return fmt.Errorf("UI_DEV_CREATE: %w", err)
	}

	// Small delay for device to register with the kernel.
	time.Sleep(50 * time.Millisecond)

	t.fd = f
	slog.Info("uinput key device created", "path", uinputPath)
	return nil
}

func (t *KeyTarget) Type() mapping.OutputType { return mapping.OutputKey }

func (t *KeyTarget) Execute(action mapping.OutputAction, value float64) error {
	if t.fd == nil {
		return fmt.Errorf("uinput key device not initialized")
	}

	// Determine press (1) or release (0) from value.
	// value > 0.5 = press, value <= 0.5 = release.
	val := int32(1)
	if value <= 0.5 {
		val = 0
	}

	for _, keyName := range action.Keys {
		code, ok := ResolveKeyCode(keyName)
		if !ok {
			slog.Warn("unknown key code", "key", keyName)
			continue
		}
		if err := t.emitKey(code, val); err != nil {
			return err
		}
	}
	return t.sync()
}

func (t *KeyTarget) emitKey(code uint16, value int32) error {
	return t.writeEvent(evKey, code, value)
}

func (t *KeyTarget) sync() error {
	return t.writeEvent(evSyn, synReport, 0)
}

func (t *KeyTarget) writeEvent(typ, code uint16, value int32) error {
	now := time.Now()
	ev := inputEvent{
		Sec:  now.Unix(),
		Usec: int64(now.Nanosecond() / 1000),
		Type: typ,
		Code: code,
		Val:  value,
	}
	buf := make([]byte, inputEventSize)
	binary.LittleEndian.PutUint64(buf[0:8], uint64(ev.Sec))
	binary.LittleEndian.PutUint64(buf[8:16], uint64(ev.Usec))
	binary.LittleEndian.PutUint16(buf[16:18], ev.Type)
	binary.LittleEndian.PutUint16(buf[18:20], ev.Code)
	binary.LittleEndian.PutUint32(buf[20:24], uint32(ev.Val))
	_, err := t.fd.Write(buf)
	return err
}

func (t *KeyTarget) Close() error {
	if t.fd == nil {
		return nil
	}
	ioctl(t.fd, uiDevDestroy, 0)
	return t.fd.Close()
}

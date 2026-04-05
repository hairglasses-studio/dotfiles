//go:build windows

package output

import (
	"fmt"
	"strconv"
	"sync"
	"syscall"
	"unsafe"

	"github.com/hairglasses-studio/mapping"
)

var (
	winmm           = syscall.NewLazyDLL("winmm.dll")
	midiOutOpen     = winmm.NewProc("midiOutOpen")
	midiOutShortMsg = winmm.NewProc("midiOutShortMsg")
	midiOutClose    = winmm.NewProc("midiOutClose")
)

// MIDITarget sends MIDI messages via Windows winmm.dll.
type MIDITarget struct {
	mu     sync.Mutex
	handle uintptr
	opened bool
}

func NewMIDITarget() *MIDITarget { return &MIDITarget{} }

func (t *MIDITarget) Type() mapping.OutputType { return mapping.OutputMIDI }

func (t *MIDITarget) open() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.opened {
		return nil
	}

	// Open default MIDI output device (device 0).
	ret, _, _ := midiOutOpen.Call(
		uintptr(unsafe.Pointer(&t.handle)),
		0,                          // device ID
		0,                          // callback
		0,                          // instance
		0,                          // flags
	)
	if ret != 0 {
		return fmt.Errorf("midiOutOpen failed: %d", ret)
	}
	t.opened = true
	return nil
}

func (t *MIDITarget) Execute(action mapping.OutputAction, value float64) error {
	if err := t.open(); err != nil {
		return err
	}

	channel := action.Port
	if channel < 0 || channel > 15 {
		channel = 0
	}

	switch {
	case action.Address != "":
		// CC message
		cc, err := strconv.Atoi(action.Address)
		if err != nil {
			return fmt.Errorf("invalid CC number: %s", action.Address)
		}
		ccVal := int(value * 127)
		if ccVal > 127 {
			ccVal = 127
		}
		return t.shortMsg(0xB0|byte(channel), byte(cc), byte(ccVal))

	case len(action.Keys) > 0:
		// Note message
		note, err := strconv.Atoi(action.Keys[0])
		if err != nil {
			return fmt.Errorf("invalid note: %s", action.Keys[0])
		}
		if value > 0.5 {
			vel := int(value * 127)
			if vel > 127 {
				vel = 127
			}
			return t.shortMsg(0x90|byte(channel), byte(note), byte(vel))
		}
		return t.shortMsg(0x80|byte(channel), byte(note), 0)
	}

	return nil
}

func (t *MIDITarget) shortMsg(status, data1, data2 byte) error {
	msg := uintptr(status) | uintptr(data1)<<8 | uintptr(data2)<<16
	t.mu.Lock()
	defer t.mu.Unlock()
	ret, _, _ := midiOutShortMsg.Call(t.handle, msg)
	if ret != 0 {
		return fmt.Errorf("midiOutShortMsg failed: %d", ret)
	}
	return nil
}

func (t *MIDITarget) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.opened {
		midiOutClose.Call(t.handle)
		t.opened = false
	}
	return nil
}

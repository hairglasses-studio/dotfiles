//go:build linux

package output

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/hairglasses-studio/mapping"
)

// MIDITarget sends MIDI messages via ALSA raw MIDI devices on Linux.
type MIDITarget struct {
	mu    sync.Mutex
	conns map[string]*os.File // device path -> fd
}

// NewMIDITarget creates an ALSA MIDI output target.
func NewMIDITarget() *MIDITarget {
	return &MIDITarget{conns: make(map[string]*os.File)}
}

func (t *MIDITarget) Type() mapping.OutputType { return mapping.OutputMIDI }

func (t *MIDITarget) Execute(action mapping.OutputAction, value float64) error {
	dev, err := t.getDevice()
	if err != nil {
		return err
	}

	channel := uint8(action.Port) // reuse Port field for MIDI channel (0-15)
	if channel > 15 {
		channel = 0
	}

	// Determine MIDI message type from the action fields.
	if action.Address != "" {
		// CC message: address = controller number as string, value = scaled 0-127
		cc := uint8(0)
		fmt.Sscanf(action.Address, "%d", &cc)
		val := uint8(clampMIDI(value))
		return t.sendCC(dev, channel, cc, val)
	}

	// Note on/off: use Keys[0] as note number
	if len(action.Keys) > 0 {
		note := uint8(0)
		fmt.Sscanf(action.Keys[0], "%d", &note)
		vel := uint8(clampMIDI(value))
		if vel > 0 {
			return t.sendNoteOn(dev, channel, note, vel)
		}
		return t.sendNoteOff(dev, channel, note)
	}

	return fmt.Errorf("midi_out: no controller or note specified")
}

func (t *MIDITarget) getDevice() (*os.File, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Return first open device.
	for _, f := range t.conns {
		return f, nil
	}

	// Auto-discover first available MIDI output.
	matches, _ := filepath.Glob("/dev/snd/midiC*D*")
	if len(matches) == 0 {
		return nil, fmt.Errorf("no ALSA MIDI devices found")
	}

	path := matches[0]
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("open MIDI device %s: %w", path, err)
	}
	t.conns[path] = f
	slog.Info("MIDI output opened", "path", path)
	return f, nil
}

func (t *MIDITarget) sendCC(f *os.File, channel, controller, value uint8) error {
	_, err := f.Write([]byte{0xB0 | channel, controller, value})
	return err
}

func (t *MIDITarget) sendNoteOn(f *os.File, channel, note, velocity uint8) error {
	_, err := f.Write([]byte{0x90 | channel, note, velocity})
	return err
}

func (t *MIDITarget) sendNoteOff(f *os.File, channel, note uint8) error {
	_, err := f.Write([]byte{0x80 | channel, note, 0})
	return err
}

func (t *MIDITarget) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, f := range t.conns {
		f.Close()
	}
	t.conns = nil
	return nil
}

func clampMIDI(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 127 {
		return 127
	}
	return v
}

//go:build !(linux || darwin || windows)

package output

import (
	"fmt"

	"github.com/hairglasses-studio/mapitall/internal/mapping"
)

// MIDITarget is a stub for platforms without ALSA MIDI support.
type MIDITarget struct{}

func NewMIDITarget() *MIDITarget               { return &MIDITarget{} }
func (t *MIDITarget) Type() mapping.OutputType { return mapping.OutputMIDI }
func (t *MIDITarget) Execute(action mapping.OutputAction, value float64) error {
	return fmt.Errorf("midi_out not implemented for this platform")
}
func (t *MIDITarget) Close() error { return nil }

//go:build !(linux || darwin || windows)

package output

import (
	"fmt"

	"github.com/hairglasses-studio/mapitall/internal/mapping"
)

// MovementTarget is a stub for platforms without uinput support.
type MovementTarget struct{}

func NewMovementTarget() *MovementTarget           { return &MovementTarget{} }
func (t *MovementTarget) Type() mapping.OutputType { return mapping.OutputMovement }
func (t *MovementTarget) Execute(action mapping.OutputAction, value float64) error {
	return fmt.Errorf("movement output not implemented for this platform")
}
func (t *MovementTarget) Close() error { return nil }

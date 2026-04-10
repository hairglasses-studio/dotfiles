//go:build !(linux || darwin || windows)

package output

import (
	"fmt"

	"github.com/hairglasses-studio/mapitall/internal/mapping"
)

// KeyTarget is a stub for platforms without uinput support.
type KeyTarget struct{}

func NewKeyTarget() *KeyTarget                { return &KeyTarget{} }
func (t *KeyTarget) Type() mapping.OutputType { return mapping.OutputKey }
func (t *KeyTarget) Execute(action mapping.OutputAction, _ float64) error {
	return fmt.Errorf("key output not implemented for this platform")
}
func (t *KeyTarget) Close() error { return nil }

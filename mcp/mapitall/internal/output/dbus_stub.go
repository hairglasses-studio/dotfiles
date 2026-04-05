//go:build !linux

package output

import (
	"fmt"

	"github.com/hairglasses-studio/mapping"
)

// DBusTarget is a no-op on non-Linux platforms.
type DBusTarget struct{}

func NewDBusTarget() *DBusTarget                                                  { return &DBusTarget{} }
func (t *DBusTarget) Type() mapping.OutputType                                    { return mapping.OutputDBus }
func (t *DBusTarget) Execute(action mapping.OutputAction, value float64) error     { return fmt.Errorf("dbus not available on this platform") }
func (t *DBusTarget) Close() error                                                { return nil }

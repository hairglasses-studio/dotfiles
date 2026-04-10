//go:build linux

package output

import (
	"fmt"
	"log/slog"

	"github.com/hairglasses-studio/mapitall/internal/mapping"

	"github.com/godbus/dbus/v5"
)

// DBusTarget calls D-Bus methods on the session or system bus.
type DBusTarget struct {
	session *dbus.Conn
}

// NewDBusTarget creates a D-Bus output target.
func NewDBusTarget() *DBusTarget {
	t := &DBusTarget{}
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		slog.Warn("D-Bus session bus connection failed (dbus output disabled)", "error", err)
		return t
	}
	t.session = conn
	return t
}

func (t *DBusTarget) Type() mapping.OutputType { return mapping.OutputDBus }

func (t *DBusTarget) Execute(action mapping.OutputAction, value float64) error {
	if t.session == nil {
		return fmt.Errorf("D-Bus not connected")
	}

	// Exec format: ["dest.service.Name", "method.Name", optional args...]
	// Example: ["org.mpris.MediaPlayer2.Player", "PlayPause"]
	// Or:      ["org.freedesktop.DBus", "ListNames"]
	if len(action.Exec) < 2 {
		return fmt.Errorf("dbus: need at least [destination, method]")
	}

	dest := action.Exec[0]
	method := action.Exec[1]

	// Build object path from destination (convention: /dest/path with dots → slashes).
	objPath := dbus.ObjectPath("/" + dotToSlash(dest))

	// Use the destination as the interface too (common pattern for MPRIS etc).
	call := t.session.Object(dest, objPath).Call(dest+"."+method, 0)
	if call.Err != nil {
		return fmt.Errorf("dbus call %s.%s: %w", dest, method, call.Err)
	}
	return nil
}

func dotToSlash(s string) string {
	out := make([]byte, len(s))
	for i := range s {
		if s[i] == '.' {
			out[i] = '/'
		} else {
			out[i] = s[i]
		}
	}
	return string(out)
}

func (t *DBusTarget) Close() error {
	if t.session != nil {
		return t.session.Close()
	}
	return nil
}

//go:build dev

package web

import (
	"io/fs"
	"os"
)

// GetStaticFS returns a dev file system pointing to web/dist.
// In dev mode, this returns nil so the server knows to proxy to Vite.
func GetStaticFS() (fs.FS, error) {
	// Check if dist exists
	if _, err := os.Stat("web/dist"); err != nil {
		return nil, err
	}
	return os.DirFS("web/dist"), nil
}

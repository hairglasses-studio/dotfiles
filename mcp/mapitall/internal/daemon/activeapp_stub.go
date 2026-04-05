//go:build !(linux || darwin)

package daemon

import "context"

// startActiveAppTracker is a no-op on unsupported platforms.
func (d *Daemon) startActiveAppTracker(_ context.Context) {}

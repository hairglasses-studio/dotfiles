//go:build linux

package service

import (
	"net"
	"os"
)

// NotifyReady sends READY=1 to systemd via the notify socket.
// This implements the sd_notify protocol without requiring libsystemd (pure Go).
func NotifyReady() {
	sdNotify("READY=1")
}

// NotifyStopping sends STOPPING=1 to systemd.
func NotifyStopping() {
	sdNotify("STOPPING=1")
}

func sdNotify(state string) {
	addr := os.Getenv("NOTIFY_SOCKET")
	if addr == "" {
		return
	}

	conn, err := net.Dial("unixgram", addr)
	if err != nil {
		return
	}
	defer conn.Close()
	_, _ = conn.Write([]byte(state))
}

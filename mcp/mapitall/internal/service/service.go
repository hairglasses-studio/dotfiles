//go:build !linux

package service

// NotifyReady signals the init system that the service is ready.
// No-op on non-Linux platforms.
func NotifyReady() {}

// NotifyStopping signals the init system that the service is shutting down.
// No-op on non-Linux platforms.
func NotifyStopping() {}

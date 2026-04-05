//go:build linux

package output

import (
	"os"
	"syscall"
	"unsafe"
)

func ioctl(f *os.File, request, arg uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), request, arg)
	if errno != 0 {
		return errno
	}
	return nil
}

func ioctlPtr(f *os.File, request uintptr, arg unsafe.Pointer) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), request, uintptr(arg))
	if errno != 0 {
		return errno
	}
	return nil
}

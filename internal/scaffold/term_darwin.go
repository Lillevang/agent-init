//go:build darwin

package scaffold

import (
	"os"
	"syscall"
	"unsafe"
)

func isTerminal(file *os.File) bool {
	var termios syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), syscall.TIOCGETA, uintptr(unsafe.Pointer(&termios)))
	return errno == 0
}

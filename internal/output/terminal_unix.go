//go:build !windows

package output

import (
	"os"
	"syscall"
	"unsafe"
)

// detectTerminalWidth uses the TIOCGWINSZ ioctl to get the terminal size.
func detectTerminalWidth() int {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	var ws winsize
	fd := os.Stdout.Fd()
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&ws)),
	)
	if errno != 0 {
		return 0
	}
	return int(ws.Col)
}

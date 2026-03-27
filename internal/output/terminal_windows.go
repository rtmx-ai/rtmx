//go:build windows

package output

// detectTerminalWidth returns 0 on Windows (falls back to DefaultTerminalWidth).
// A future implementation could use GetConsoleScreenBufferInfo via the Windows API.
func detectTerminalWidth() int {
	return 0
}

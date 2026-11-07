//go:build windows

package broker

func processPrivileged() bool {
	// Service registration still expects an elevated shell on Windows.
	return false
}

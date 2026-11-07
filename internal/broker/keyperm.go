// Package broker installs and runs the verifier service (REQ-MONO-069).
package broker

import (
	"fmt"
	"os"
	"runtime"
)

// EnsurePrivateKeyMode600 sets the private key file to mode 0600 so only the
// owning verifier account can read it (REQ-MONO-069ba).
// On Windows, POSIX mode bits are not authoritative; isolation is enforced via
// DefaultWindowsKeyACL (REQ-MONO-069bc).
func EnsurePrivateKeyMode600(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Mode().Perm() != 0o600 {
		return fmt.Errorf("broker: key mode %o, want 0600", info.Mode().Perm())
	}
	return nil
}

// PrivateKeyReadableByOthers reports whether group or other have any access.
// A developer account that is not the key owner must not be able to read it.
// On Windows this always returns false; use DefaultWindowsKeyACL instead.
func PrivateKeyReadableByOthers(path string) (bool, error) {
	if runtime.GOOS == "windows" {
		return false, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	perm := info.Mode().Perm()
	return perm&0o077 != 0, nil
}

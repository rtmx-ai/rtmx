//go:build !windows

package broker

import "os"

func processPrivileged() bool {
	return os.Geteuid() == 0
}

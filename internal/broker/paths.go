package broker

import (
	"path/filepath"
	"runtime"
)

// PackagePaths are the on-disk locations for the verifier that ships inside
// the rtmx package (REQ-MONO-069b / 069j). Homebrew/Cellar may stage binaries
// under a prefix; install copies or links them into these system paths.
type PackagePaths struct {
	// VerifyBinary is the daemon executable (libexec).
	VerifyBinary string
	// KeyPath is the private signing key (mode 0600, verifier-owned).
	KeyPath string
	// SocketPath is the unix domain socket clients dial (non-Windows).
	SocketPath string
	// LaunchdPlist is the macOS LaunchDaemon path.
	LaunchdPlist string
	// SystemdUnit is the Linux systemd unit path.
	SystemdUnit string
	// WindowsServiceJSON is the Windows service contract path (share).
	WindowsServiceJSON string
	// VerifierAccount is the dedicated OS account that runs the broker.
	VerifierAccount string
}

// DefaultPackagePaths returns the production layout for the current OS.
func DefaultPackagePaths() PackagePaths {
	p := PackagePaths{
		KeyPath:         "/var/lib/rtmx/verify.key",
		SocketPath:      DefaultBrokerSocket(),
		VerifierAccount: "rtmx-verify",
	}
	switch runtime.GOOS {
	case "darwin":
		p.VerifyBinary = "/usr/local/libexec/rtmx/verify"
		p.LaunchdPlist = "/Library/LaunchDaemons/com.rtmx.verify.plist"
		p.SocketPath = "/var/run/rtmx/verify.sock"
	case "windows":
		p.VerifyBinary = `C:\Program Files\RTMX\verify.exe`
		p.KeyPath = `C:\ProgramData\RTMX\verify.key`
		p.VerifierAccount = `NT SERVICE\rtmx`
		p.SocketPath = `\\.\pipe\rtmx-verify`
	default: // linux and others (deb/rpm place verify in bindir)
		p.VerifyBinary = "/usr/bin/verify"
		p.SystemdUnit = "/etc/systemd/system/rtmx-verify.service"
	}
	return p
}

// DefaultBrokerSocket is the socket the install enables at boot.
func DefaultBrokerSocket() string {
	if runtime.GOOS == "darwin" {
		return "/var/run/rtmx/verify.sock"
	}
	if runtime.GOOS == "windows" {
		return `\\.\pipe\rtmx-verify`
	}
	return "/run/rtmx/verify.sock"
}

// StagedPaths returns Cellar/prefix-relative staging locations used by the
// Homebrew/nfpm package before privileged install.
func StagedPaths(prefix string) PackagePaths {
	return PackagePaths{
		VerifyBinary:       filepath.Join(prefix, "libexec", "rtmx", "verify"),
		KeyPath:            DefaultPackagePaths().KeyPath,
		SocketPath:         DefaultPackagePaths().SocketPath,
		LaunchdPlist:       filepath.Join(prefix, "share", "rtmx", "com.rtmx.verify.plist"),
		SystemdUnit:        filepath.Join(prefix, "share", "rtmx", "rtmx-verify.service"),
		WindowsServiceJSON: filepath.Join(prefix, "share", "rtmx", "verify-service.json"),
		VerifierAccount:    DefaultPackagePaths().VerifierAccount,
	}
}

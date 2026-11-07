package broker

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

// EnsureOptions controls automatic broker enablement (REQ-MONO-069j).
type EnsureOptions struct {
	SocketPath string
	Prefix     string // Homebrew Cellar / package prefix for --prefix
	Binary     string // packaged verify binary for --binary
	MarkerDir  string
	// Elevate is optional; defaults to ElevateOS.
	Elevate ElevateFunc
	// Reachable overrides SocketReachable in tests.
	Reachable func(socketPath string) bool
	// ExecRTMX runs `rtmx broker install|upgrade` when already privileged.
	// Defaults to exec of os.Executable().
	ExecRTMX func(args []string) error
	// PreferUpgrade uses `broker upgrade` instead of `install`.
	PreferUpgrade bool
	// SkipElevate refuses elevation (tests / ConnectWithoutAdmin paths).
	SkipElevate bool
}

var ensureOnce sync.Mutex

// EnsureBrokerInstalled enables the packaged broker when the socket is down.
// Elevation happens at most once per process and is suppressed for
// attemptCooldown after a failed attempt (REQ-MONO-069j).
func EnsureBrokerInstalled(opts EnsureOptions) error {
	ensureOnce.Lock()
	defer ensureOnce.Unlock()

	sock := opts.SocketPath
	if sock == "" {
		sock = DefaultBrokerSocket()
	}
	reach := opts.Reachable
	if reach == nil {
		reach = SocketReachable
	}
	if reach(sock) {
		return nil
	}
	if opts.SkipElevate {
		return fmt.Errorf("verify: local broker not reachable at %s; run `%s`", sock, BrokerInstallHint)
	}

	marker := AttemptMarkerPath(opts.MarkerDir)
	if recentAttempt(marker) {
		return fmt.Errorf("verify: broker still unreachable at %s after a recent install attempt; run `%s` or wait and retry", sock, BrokerInstallHint)
	}

	args := []string{"broker", "install"}
	if opts.PreferUpgrade {
		args[1] = "upgrade"
	}
	if opts.Prefix != "" {
		args = append(args, "--prefix", opts.Prefix)
	}
	if opts.Binary != "" {
		args = append(args, "--binary", opts.Binary)
	}

	elevate := opts.Elevate
	if elevate == nil {
		elevate = ElevateOS
	}
	run := opts.ExecRTMX
	if run == nil {
		run = execSelf
	}

	var err error
	if IsPrivileged() {
		err = run(args)
	} else {
		exe, exeErr := os.Executable()
		if exeErr != nil {
			markAttempt(marker)
			return fmt.Errorf("verify: locate rtmx for elevated broker install: %w", exeErr)
		}
		err = elevate(append([]string{exe}, args...))
	}
	markAttempt(marker)
	if err != nil {
		return fmt.Errorf("verify: automatic broker enable failed: %w (hint: `%s`)", err, BrokerInstallHint)
	}

	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		if reach(sock) {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("verify: broker installed but not reachable at %s; check `rtmx broker status`", sock)
}

func execSelf(args []string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

package broker

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ElevateFunc re-executes argv[0] with argv[1:] under administrator privileges.
// Tests inject a stub; production uses ElevateOS.
type ElevateFunc func(argv []string) error

// ElevateOS is the platform elevation entry point (sudo / osascript / pkexec).
var ElevateOS ElevateFunc = elevateOS

// IsPrivileged reports whether the current process can write system broker paths.
func IsPrivileged() bool {
	if processPrivileged() {
		return true
	}
	u, err := user.Current()
	if err != nil {
		return false
	}
	return u.Uid == "0"
}

func elevateOS(argv []string) error {
	if len(argv) == 0 {
		return fmt.Errorf("broker elevate: empty argv")
	}
	switch runtime.GOOS {
	case "darwin":
		return elevateDarwin(argv)
	case "windows":
		return fmt.Errorf("broker elevate: on Windows run an elevated shell: %s", strings.Join(argv, " "))
	default:
		return elevateLinux(argv)
	}
}

func elevateDarwin(argv []string) error {
	quoted := shellJoin(argv)
	script := fmt.Sprintf(`do shell script %q with administrator privileges`, quoted)
	cmd := exec.Command("osascript", "-e", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func elevateLinux(argv []string) error {
	if path, err := exec.LookPath("pkexec"); err == nil {
		cmd := exec.Command(path, argv...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	if path, err := exec.LookPath("sudo"); err == nil {
		args := append([]string{"-n"}, argv...)
		cmd := exec.Command(path, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
		args = append([]string{"--"}, argv...)
		cmd = exec.Command(path, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}
	return fmt.Errorf("broker elevate: neither pkexec nor sudo found; run `%s` as root", BrokerInstallHint)
}

func shellJoin(argv []string) string {
	parts := make([]string, len(argv))
	for i, a := range argv {
		parts[i] = "'" + strings.ReplaceAll(a, "'", `'"'"'`) + "'"
	}
	return strings.Join(parts, " ")
}

// AttemptMarkerPath is where EnsureBrokerInstalled records a recent elevate try.
func AttemptMarkerPath(dir string) string {
	if dir == "" {
		dir = defaultMarkerDir()
	}
	return filepath.Join(dir, "ensure-broker.attempt")
}

func defaultMarkerDir() string {
	if runtime.GOOS == "windows" {
		if d := os.Getenv("LOCALAPPDATA"); d != "" {
			return filepath.Join(d, "rtmx")
		}
	}
	if d, err := os.UserCacheDir(); err == nil && d != "" {
		return filepath.Join(d, "rtmx")
	}
	return os.TempDir()
}

const attemptCooldown = 5 * time.Minute

func recentAttempt(marker string) bool {
	st, err := os.Stat(marker)
	if err != nil {
		return false
	}
	return time.Since(st.ModTime()) < attemptCooldown
}

func markAttempt(marker string) {
	_ = os.MkdirAll(filepath.Dir(marker), 0o755)
	_ = os.WriteFile(marker, []byte(time.Now().UTC().Format(time.RFC3339)+"\n"), 0o644)
}

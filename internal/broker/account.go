package broker

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// CommandRunner runs external provisioning commands (dscl, useradd, …).
// Tests inject a stub.
type CommandRunner func(name string, args ...string) error

// RunCommand is the production CommandRunner.
var RunCommand CommandRunner = func(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// EnsureVerifierAccount creates the dedicated OS account/group when missing.
// No-op on Windows (service account is declared in the service contract).
func EnsureVerifierAccount(account string) error {
	if account == "" {
		account = "rtmx-verify"
	}
	switch runtime.GOOS {
	case "windows":
		return nil
	case "darwin":
		return ensureDarwinAccount(account)
	default:
		return ensureLinuxAccount(account)
	}
}

func ensureDarwinAccount(account string) error {
	gidStr := ""
	if g, err := user.LookupGroup(account); err == nil {
		gidStr = g.Gid
	} else {
		gid, err := nextDarwinID(true)
		if err != nil {
			return err
		}
		gidStr = strconv.Itoa(gid)
		steps := [][]string{
			{"dscl", ".", "-create", "/Groups/" + account},
			{"dscl", ".", "-create", "/Groups/" + account, "PrimaryGroupID", gidStr},
			{"dscl", ".", "-create", "/Groups/" + account, "RealName", "RTMX Verify"},
		}
		for _, args := range steps {
			if err := RunCommand(args[0], args[1:]...); err != nil {
				return fmt.Errorf("broker install: create group %s: %w", account, err)
			}
		}
	}
	if _, err := user.Lookup(account); err == nil {
		return nil
	}
	uid, err := nextDarwinID(false)
	if err != nil {
		return err
	}
	steps := [][]string{
		{"dscl", ".", "-create", "/Users/" + account},
		{"dscl", ".", "-create", "/Users/" + account, "UniqueID", strconv.Itoa(uid)},
		{"dscl", ".", "-create", "/Users/" + account, "PrimaryGroupID", gidStr},
		{"dscl", ".", "-create", "/Users/" + account, "UserShell", "/usr/bin/false"},
		{"dscl", ".", "-create", "/Users/" + account, "RealName", "RTMX Verify"},
		{"dscl", ".", "-create", "/Users/" + account, "NFSHomeDirectory", "/var/lib/rtmx"},
	}
	for _, args := range steps {
		if err := RunCommand(args[0], args[1:]...); err != nil {
			return fmt.Errorf("broker install: create user %s: %w", account, err)
		}
	}
	return nil
}

func nextDarwinID(group bool) (int, error) {
	// Prefer the macOS daemon range below 500.
	for id := 260; id < 400; id++ {
		if group {
			if _, err := user.LookupGroupId(strconv.Itoa(id)); err != nil {
				return id, nil
			}
			continue
		}
		if _, err := user.LookupId(strconv.Itoa(id)); err != nil {
			return id, nil
		}
	}
	return 0, fmt.Errorf("broker install: no free uid/gid in daemon range")
}

func ensureLinuxAccount(account string) error {
	if _, err := user.LookupGroup(account); err != nil {
		if err := RunCommand("groupadd", "--system", account); err != nil {
			return fmt.Errorf("broker install: groupadd %s: %w", account, err)
		}
	}
	if _, err := user.Lookup(account); err == nil {
		return nil
	}
	if err := RunCommand("useradd", "--system", "--gid", account, "--home-dir", "/var/lib/rtmx",
		"--shell", "/usr/sbin/nologin", "--comment", "RTMX Verify", account); err != nil {
		return fmt.Errorf("broker install: useradd %s: %w", account, err)
	}
	return nil
}

// EnsureRuntimeDirs creates socket and work directories owned by the verifier.
func EnsureRuntimeDirs(paths PackagePaths) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	dirs := []string{
		filepath.Dir(paths.SocketPath),
		"/var/lib/rtmx",
		"/var/lib/rtmx/work",
		filepath.Dir(paths.KeyPath),
	}
	seen := map[string]bool{}
	for _, d := range dirs {
		if d == "" || d == "." || d == "/" || seen[d] {
			continue
		}
		seen[d] = true
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("broker install: mkdir %s: %w", d, err)
		}
	}
	return ChownVerifierPaths(paths)
}

// ChownVerifierPaths sets ownership of key/runtime dirs to the verifier account.
func ChownVerifierPaths(paths PackagePaths) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	account := paths.VerifierAccount
	if account == "" {
		account = "rtmx-verify"
	}
	u, err := user.Lookup(account)
	if err != nil {
		return fmt.Errorf("broker install: lookup %s for chown: %w", account, err)
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return err
	}
	targets := []string{
		"/var/lib/rtmx",
		"/var/lib/rtmx/work",
		paths.KeyPath,
		filepath.Dir(paths.KeyPath),
		filepath.Dir(paths.SocketPath),
	}
	for _, t := range targets {
		if t == "" || strings.Contains(t, `\`) {
			continue
		}
		if _, err := os.Stat(t); err != nil {
			continue
		}
		if err := os.Chown(t, uid, gid); err != nil {
			return fmt.Errorf("broker install: chown %s: %w", t, err)
		}
	}
	return nil
}

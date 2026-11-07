package broker

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// InstallLayout is a DESTDIR-style layout used by rtmx broker install/upgrade.
// Root defaults to filesystem root; tests pass a temp dir.
type InstallLayout struct {
	Root         string // empty = /
	SourceBinary string // packaged verify binary to copy
	Command      []string
	Paths        PackagePaths
}

// InstallPlan is the privileged enablement step (REQ-MONO-069j).
type InstallPlan struct {
	Layout InstallLayout
}

func resolvePath(root, abs string) string {
	if root == "" {
		return abs
	}
	// DESTDIR-style staging: strip volume (C:) and leading separators so
	// filepath.Join(root, ...) never produces root\C:\... on Windows.
	path := abs
	if vol := filepath.VolumeName(path); vol != "" {
		path = path[len(vol):]
	} else if len(path) >= 2 && ((path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z')) && path[1] == ':' {
		// Drive letter when filepath is unix-flavored (local tests / cross builds).
		path = path[2:]
	}
	path = strings.ReplaceAll(path, `\`, "/")
	path = strings.TrimLeft(path, "/")
	return filepath.Join(root, filepath.FromSlash(path))
}

// RenderLaunchdPlist returns a LaunchDaemon plist with resolved paths.
func RenderLaunchdPlist(p PackagePaths, command []string) string {
	cmd := append([]string{}, command...)
	if len(cmd) == 0 {
		cmd = []string{"go", "test", "./...", "-count=1"}
	}
	args := []string{p.VerifyBinary, "--socket", p.SocketPath, "--key", p.KeyPath, "--workdir", "/var/lib/rtmx/work"}
	for _, c := range cmd {
		args = append(args, "--command", c)
	}
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.rtmx.verify</string>
	<key>UserName</key>
	<string>rtmx-verify</string>
	<key>GroupName</key>
	<string>rtmx-verify</string>
	<key>ProgramArguments</key>
	<array>
`)
	for _, a := range args {
		b.WriteString("\t\t<string>")
		b.WriteString(html.EscapeString(a))
		b.WriteString("</string>\n")
	}
	b.WriteString(`	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
</dict>
</plist>
`)
	return b.String()
}

// RenderSystemdUnit returns a systemd unit with resolved paths.
func RenderSystemdUnit(p PackagePaths, command []string) string {
	cmd := append([]string{}, command...)
	if len(cmd) == 0 {
		cmd = []string{"go", "test", "./...", "-count=1"}
	}
	exec := fmt.Sprintf("%s --socket %s --key %s --workdir /var/lib/rtmx/work", p.VerifyBinary, p.SocketPath, p.KeyPath)
	for _, c := range cmd {
		exec += " --command " + c
	}
	account := p.VerifierAccount
	if account == "" {
		account = "rtmx-verify"
	}
	return fmt.Sprintf(`[Unit]
Description=RTMX verification broker
After=network.target

[Service]
Type=simple
User=%s
Group=%s
StateDirectory=rtmx
ExecStart=%s
Restart=on-failure
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
`, account, account, exec)
}

// Apply writes key, binary, and service unit under Layout.Root (testable without sudo).
// Real machines use Root="" so paths are absolute system locations.
func (p InstallPlan) Apply() error {
	paths := p.Layout.Paths
	if paths.VerifyBinary == "" {
		paths = DefaultPackagePaths()
		p.Layout.Paths = paths
	}
	if p.Layout.SourceBinary == "" {
		return fmt.Errorf("broker install: source verify binary is required (shipped in the rtmx package)")
	}
	raw, err := os.ReadFile(p.Layout.SourceBinary)
	if err != nil {
		return fmt.Errorf("broker install: read package binary: %w", err)
	}
	binDest := resolvePath(p.Layout.Root, paths.VerifyBinary)
	if err := os.MkdirAll(filepath.Dir(binDest), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(binDest, raw, 0o755); err != nil {
		return err
	}

	keyDest := resolvePath(p.Layout.Root, paths.KeyPath)
	if err := os.MkdirAll(filepath.Dir(keyDest), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(keyDest); os.IsNotExist(err) {
		_, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}
		if err := os.WriteFile(keyDest, priv, 0o600); err != nil {
			return err
		}
	}
	if err := EnsurePrivateKeyMode600(keyDest); err != nil {
		return err
	}

	switch runtime.GOOS {
	case "darwin":
		body := RenderLaunchdPlist(paths, p.Layout.Command)
		plist := resolvePath(p.Layout.Root, paths.LaunchdPlist)
		if err := os.MkdirAll(filepath.Dir(plist), 0o755); err != nil {
			return err
		}
		return os.WriteFile(plist, []byte(body), 0o644)
	case "windows":
		dst := paths.WindowsServiceJSON
		if dst == "" {
			dst = filepath.Join(filepath.Dir(paths.VerifyBinary), "verify-service.json")
		}
		dst = resolvePath(p.Layout.Root, dst)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		contract := fmt.Sprintf("{\"name\":\"rtmx-verify\",\"executable\":%q,\"serviceAccount\":%q,\"startType\":\"automatic\"}\n",
			paths.VerifyBinary, paths.VerifierAccount)
		return os.WriteFile(dst, []byte(contract), 0o644)
	default:
		unit := RenderSystemdUnit(paths, p.Layout.Command)
		dst := resolvePath(p.Layout.Root, paths.SystemdUnit)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, []byte(unit), 0o644)
	}
}

// BrokerInstallHint is the user-facing message when the broker is missing.
const BrokerInstallHint = "sudo rtmx broker install"

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var (
	loopInstallDry   bool
	loopInstallForce bool
	loopInstallGOOS  string

	// loopEnable would load the unit. Install never calls it.
	loopEnable func(name string, args ...string) error
)

var loopInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Write a host unit for rtmx loop without enabling it",
	Long: `Write a project-local unit so the delivery loop can survive a terminal close.

macOS: .rtmx/loop/ai.rtmx.loop.plist (LaunchAgent). Enable yourself with launchctl.
Linux: .rtmx/loop/rtmx-loop.service (systemd user unit). Enable yourself with systemctl --user.

This command does not call launchctl or systemctl.`,
	RunE: runLoopInstall,
}

func init() {
	loopInstallCmd.Flags().BoolVar(&loopInstallDry, "dry-run", false, "print the unit and write nothing")
	loopInstallCmd.Flags().BoolVar(&loopInstallForce, "force", false, "overwrite an existing unit")
	loopCmd.AddCommand(loopInstallCmd)
}

func runLoopInstall(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return err
	}
	goos := loopInstallGOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	path, body, enable, err := renderLoopUnit(abs, goos)
	if err != nil {
		return err
	}
	if loopInstallDry {
		cmd.Println(body)
		cmd.Printf("Enable (not run): %s\n", enable)
		cmd.Println("Dry run — no changes written")
		return nil
	}
	if _, err := os.Stat(path); err == nil && !loopInstallForce {
		return fmt.Errorf("unit already exists at %s; use --force to overwrite", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	_ = EnsureLoopGitignore(filepath.Join(abs, ".rtmx"))
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return err
	}
	cmd.Printf("Wrote %s\n", path)
	cmd.Printf("Enable (not run): %s\n", enable)
	cmd.Println("This command does not load or enable the unit.")
	return nil
}

func renderLoopUnit(project, goos string) (path, body, enable string, err error) {
	switch goos {
	case "darwin":
		path = filepath.Join(project, ".rtmx", "loop", "ai.rtmx.loop.plist")
		body = renderLaunchAgent(project)
		enable = fmt.Sprintf("launchctl bootstrap gui/$(id -u) %s", path)
		return path, body, enable, nil
	case "linux":
		path = filepath.Join(project, ".rtmx", "loop", "rtmx-loop.service")
		body = renderSystemdUnit(project)
		enable = fmt.Sprintf("systemctl --user link %s && systemctl --user enable rtmx-loop.service", path)
		return path, body, enable, nil
	default:
		return "", "", "", fmt.Errorf("rtmx loop install does not support %s", goos)
	}
}

func renderLaunchAgent(project string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>ai.rtmx.loop</string>
  <key>WorkingDirectory</key>
  <string>%s</string>
  <key>ProgramArguments</key>
  <array>
    <string>rtmx</string>
    <string>loop</string>
  </array>
  <key>RunAtLoad</key>
  <false/>
</dict>
</plist>
`, xmlEscape(project))
}

func renderSystemdUnit(project string) string {
	return fmt.Sprintf(`[Unit]
Description=RTMX delivery loop

[Service]
Type=simple
WorkingDirectory=%s
ExecStart=rtmx loop

[Install]
WantedBy=default.target
`, project)
}

func xmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

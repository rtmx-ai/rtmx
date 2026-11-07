package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/broker"
	"github.com/spf13/cobra"
)

var (
	brokerSourceBinary string
	brokerPrefix       string
)

var brokerCmd = &cobra.Command{
	Use:   "broker",
	Short: "Install and manage the packaged verification broker",
	Long: `The verification broker ships inside the rtmx package. Enabling it is a
one-time administrator step (REQ-MONO-069j): create the verifier account,
install the LaunchDaemon/systemd unit, and place the signing key.

  sudo rtmx broker install
  sudo rtmx broker upgrade
  rtmx broker status

rtmx verify never prompts for administrator credentials.`,
}

var brokerInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Enable the packaged broker (requires administrator)",
	RunE:  runBrokerInstall,
}

var brokerUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Replace the broker binary from the rtmx package and reload the service",
	RunE:  runBrokerInstall, // same Apply path; reloads after
}

var brokerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show whether the packaged broker is reachable",
	RunE: func(cmd *cobra.Command, _ []string) error {
		sock := broker.DefaultBrokerSocket()
		if broker.SocketReachable(sock) {
			cmd.Printf("broker: reachable at %s\n", sock)
			return nil
		}
		cmd.Printf("broker: not reachable at %s\n", sock)
		cmd.Printf("install with: %s\n", broker.BrokerInstallHint)
		return NewExitError(1, "broker not installed or not running")
	},
}

func init() {
	brokerInstallCmd.Flags().StringVar(&brokerSourceBinary, "binary", "", "path to packaged verify binary (default: discover next to rtmx)")
	brokerUpgradeCmd.Flags().StringVar(&brokerSourceBinary, "binary", "", "path to packaged verify binary (default: discover next to rtmx)")
	brokerCmd.PersistentFlags().StringVar(&brokerPrefix, "prefix", "", "package prefix that stages libexec/share (Homebrew Cellar)")
	brokerCmd.AddCommand(brokerInstallCmd)
	brokerCmd.AddCommand(brokerUpgradeCmd)
	brokerCmd.AddCommand(brokerStatusCmd)
	rootCmd.AddCommand(brokerCmd)
}

func runBrokerInstall(cmd *cobra.Command, _ []string) error {
	src, err := resolvePackagedVerifyBinary(brokerSourceBinary)
	if err != nil {
		return err
	}
	paths := broker.DefaultPackagePaths()
	plan := broker.InstallPlan{Layout: broker.InstallLayout{
		SourceBinary: src,
		Paths:        paths,
	}}
	if err := plan.Apply(); err != nil {
		return fmt.Errorf("%w\n(hint: run as root/administrator: %s)", err, broker.BrokerInstallHint)
	}
	if err := enableBrokerService(paths); err != nil {
		cmd.Printf("files installed; service enable warning: %v\n", err)
	}
	cmd.Printf("broker installed: binary=%s key=%s socket=%s\n", paths.VerifyBinary, paths.KeyPath, paths.SocketPath)
	cmd.Println("rtmx verify will connect without prompting for administrator credentials.")
	return nil
}

func resolvePackagedVerifyBinary(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if brokerPrefix != "" {
		candidate := filepath.Join(brokerPrefix, "libexec", "rtmx", "verify")
		if runtime.GOOS == "windows" {
			candidate += ".exe"
		}
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate, nil
		}
	}
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("broker: locate rtmx executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(exe)
	candidates := []string{
		filepath.Join(dir, "verify"),
		filepath.Join(dir, "rtmx-verify"),
		filepath.Join(dir, "..", "libexec", "rtmx", "verify"),
		filepath.Join(dir, "..", "libexec", "verify"),
	}
	if runtime.GOOS == "windows" {
		for i, c := range candidates {
			candidates[i] = c + ".exe"
		}
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("broker: packaged verify binary not found next to %s; reinstall rtmx or pass --binary", exe)
}

func enableBrokerService(paths broker.PackagePaths) error {
	switch runtime.GOOS {
	case "darwin":
		if paths.LaunchdPlist == "" {
			return nil
		}
		_ = exec.Command("launchctl", "bootout", "system/"+filepath.Base(strings.TrimSuffix(filepath.Base(paths.LaunchdPlist), ".plist"))).Run()
		out, err := exec.Command("launchctl", "bootstrap", "system", paths.LaunchdPlist).CombinedOutput()
		if err != nil {
			// Older macOS used load.
			out2, err2 := exec.Command("launchctl", "load", "-w", paths.LaunchdPlist).CombinedOutput()
			if err2 != nil {
				return fmt.Errorf("launchctl: %v (%s) / %v (%s)", err, out, err2, out2)
			}
		}
		return nil
	case "windows":
		return fmt.Errorf("windows service registration: use the packaged verify-service.json with sc.exe / service manager")
	default:
		if paths.SystemdUnit == "" {
			return nil
		}
		_ = exec.Command("systemctl", "daemon-reload").Run()
		out, err := exec.Command("systemctl", "enable", "--now", filepath.Base(paths.SystemdUnit)).CombinedOutput()
		if err != nil {
			return fmt.Errorf("systemctl: %v (%s)", err, out)
		}
		return nil
	}
}

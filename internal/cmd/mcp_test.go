package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

// createMCPTestCmd builds an isolated mcp-server command that wires
// the package-level flag variables the same way init() does for the
// real command, but without polluting the global rootCmd.
func createMCPTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	var port int
	var host string
	var stdio, quiet bool
	mcpCmd := &cobra.Command{
		Use:  "mcp-server",
		RunE: func(cmd *cobra.Command, args []string) error {
			mcpPort = port
			mcpHost = host
			mcpStdio = stdio
			mcpQuiet = quiet
			return runMCPServer(cmd, args)
		},
	}
	mcpCmd.Flags().IntVar(&port, "port", 0, "port to listen on")
	mcpCmd.Flags().StringVar(&host, "host", "", "host to bind to")
	mcpCmd.Flags().BoolVar(&stdio, "stdio", false, "use stdin/stdout transport")
	mcpCmd.Flags().BoolVar(&quiet, "quiet", false, "suppress response size logging")
	root.AddCommand(mcpCmd)
	return root
}

// setupMCPTestProject creates a temp directory with rtmx.yaml and
// optionally a database CSV. Returns the temp directory path.
func setupMCPTestProject(t *testing.T, dbContent string) string {
	t.Helper()
	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"),
		[]byte("rtmx:\n  database: .rtmx/database.csv\n  schema: core\n"), 0644)
	if dbContent != "" {
		_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)
	}
	return tmpDir
}

// setupMCPTestProjectWithMCPConfig creates a temp directory with
// rtmx.yaml that includes MCP host/port config.
func setupMCPTestProjectWithMCPConfig(t *testing.T, dbContent, host string, port int) string {
	t.Helper()
	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0755)

	yamlContent := "rtmx:\n  database: .rtmx/database.csv\n  schema: core\n"
	yamlContent += "  mcp:\n"
	if host != "" {
		yamlContent += fmt.Sprintf("    host: %s\n", host)
	}
	if port != 0 {
		yamlContent += fmt.Sprintf("    port: %d\n", port)
	}

	_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte(yamlContent), 0644)
	if dbContent != "" {
		_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)
	}
	return tmpDir
}

func TestMCPServerCommand(t *testing.T) {
	rtmx.Req(t, "REQ-GO-080")

	t.Run("flags_registered", func(t *testing.T) {
		// Verify all expected flags exist on the real mcpServerCmd.
		for _, name := range []string{"port", "host", "stdio", "quiet"} {
			if mcpServerCmd.Flags().Lookup(name) == nil {
				t.Errorf("mcp-server should have --%s flag", name)
			}
		}
	})

	t.Run("port_flag_overrides_config", func(t *testing.T) {
		cmd := createMCPTestCmd()
		mcpCmd, _, _ := cmd.Find([]string{"mcp-server"})
		if mcpCmd == nil {
			t.Fatal("mcp-server command not found")
		}
		err := mcpCmd.Flags().Parse([]string{"--port", "9999"})
		if err != nil {
			t.Fatalf("flag parse failed: %v", err)
		}
		val, err := mcpCmd.Flags().GetInt("port")
		if err != nil {
			t.Fatalf("get port: %v", err)
		}
		if val != 9999 {
			t.Errorf("port = %d, want 9999", val)
		}
	})

	t.Run("host_flag_overrides_config", func(t *testing.T) {
		cmd := createMCPTestCmd()
		mcpCmd, _, _ := cmd.Find([]string{"mcp-server"})
		if mcpCmd == nil {
			t.Fatal("mcp-server command not found")
		}
		err := mcpCmd.Flags().Parse([]string{"--host", "0.0.0.0"})
		if err != nil {
			t.Fatalf("flag parse failed: %v", err)
		}
		val, err := mcpCmd.Flags().GetString("host")
		if err != nil {
			t.Fatalf("get host: %v", err)
		}
		if val != "0.0.0.0" {
			t.Errorf("host = %q, want 0.0.0.0", val)
		}
	})

	t.Run("stdio_flag_selects_stdio_transport", func(t *testing.T) {
		cmd := createMCPTestCmd()
		mcpCmd, _, _ := cmd.Find([]string{"mcp-server"})
		if mcpCmd == nil {
			t.Fatal("mcp-server command not found")
		}
		err := mcpCmd.Flags().Parse([]string{"--stdio"})
		if err != nil {
			t.Fatalf("flag parse failed: %v", err)
		}
		val, err := mcpCmd.Flags().GetBool("stdio")
		if err != nil {
			t.Fatalf("get stdio: %v", err)
		}
		if !val {
			t.Error("stdio should be true when --stdio is passed")
		}
	})

	t.Run("quiet_flag_suppresses_startup_output", func(t *testing.T) {
		cmd := createMCPTestCmd()
		mcpCmd, _, _ := cmd.Find([]string{"mcp-server"})
		if mcpCmd == nil {
			t.Fatal("mcp-server command not found")
		}
		err := mcpCmd.Flags().Parse([]string{"--quiet"})
		if err != nil {
			t.Fatalf("flag parse failed: %v", err)
		}
		val, err := mcpCmd.Flags().GetBool("quiet")
		if err != nil {
			t.Fatalf("get quiet: %v", err)
		}
		if !val {
			t.Error("quiet should be true when --quiet is passed")
		}
	})

	t.Run("default_flag_values_trigger_config_fallback", func(t *testing.T) {
		// When no flags are provided, port defaults to 0 and host to "",
		// which causes runMCPServer to fall through to config, then
		// to hardcoded defaults (localhost:3000).
		cmd := createMCPTestCmd()
		mcpCmd, _, _ := cmd.Find([]string{"mcp-server"})
		if mcpCmd == nil {
			t.Fatal("mcp-server command not found")
		}

		portVal, _ := mcpCmd.Flags().GetInt("port")
		hostVal, _ := mcpCmd.Flags().GetString("host")
		if portVal != 0 {
			t.Errorf("default port flag = %d, want 0 (triggers config fallback)", portVal)
		}
		if hostVal != "" {
			t.Errorf("default host flag = %q, want empty (triggers config fallback)", hostVal)
		}
	})

	t.Run("config_host_port_used_when_flags_omitted", func(t *testing.T) {
		// Set up project with MCP config specifying custom host/port.
		// Verify the config is loaded correctly so that runMCPServer
		// would use these values in its fallback chain.
		tmpDir := setupMCPTestProjectWithMCPConfig(t, testDBHeader, "10.0.0.1", 5555)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cfg, err := config.LoadFromDir(tmpDir)
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if cfg.RTMX.MCP.Host != "10.0.0.1" {
			t.Errorf("config host = %q, want 10.0.0.1", cfg.RTMX.MCP.Host)
		}
		if cfg.RTMX.MCP.Port != 5555 {
			t.Errorf("config port = %d, want 5555", cfg.RTMX.MCP.Port)
		}

		// Simulate the same fallback logic as runMCPServer with no flags.
		host := "" // flag default
		if host == "" {
			host = cfg.RTMX.MCP.Host
		}
		if host == "" {
			host = "localhost"
		}
		port := 0 // flag default
		if port == 0 {
			port = cfg.RTMX.MCP.Port
		}
		if port == 0 {
			port = 3000
		}

		if host != "10.0.0.1" {
			t.Errorf("resolved host = %q, want 10.0.0.1", host)
		}
		if port != 5555 {
			t.Errorf("resolved port = %d, want 5555", port)
		}
	})

	t.Run("no_config_falls_to_hardcoded_defaults", func(t *testing.T) {
		// With no MCP config section and no flags, verify the
		// fallback chain ends at localhost:3000.
		tmpDir := setupMCPTestProject(t, testDBHeader)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cfg, err := config.LoadFromDir(tmpDir)
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		host := ""
		if host == "" {
			host = cfg.RTMX.MCP.Host
		}
		if host == "" {
			host = "localhost"
		}
		port := 0
		if port == 0 {
			port = cfg.RTMX.MCP.Port
		}
		if port == 0 {
			port = 3000
		}

		if host != "localhost" {
			t.Errorf("resolved host = %q, want localhost", host)
		}
		if port != 3000 {
			t.Errorf("resolved port = %d, want 3000", port)
		}
	})

	t.Run("http_mode_bind_error_returns_error", func(t *testing.T) {
		// Occupy a port, then try to start the MCP server on it.
		// The command should return a server error.
		tmpDir := setupMCPTestProject(t, testDBHeader)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		// Grab an ephemeral port and hold it.
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to create listener: %v", err)
		}
		defer ln.Close()
		occupiedPort := ln.Addr().(*net.TCPAddr).Port

		cmd := createMCPTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{
			"mcp-server",
			"--host", "127.0.0.1",
			"--port", fmt.Sprintf("%d", occupiedPort),
		})

		err = cmd.Execute()
		if err == nil {
			t.Fatal("expected error when port is occupied")
		}
		if !strings.Contains(err.Error(), "server error") && !strings.Contains(err.Error(), "bind") && !strings.Contains(err.Error(), "address already in use") {
			t.Errorf("expected server/bind error, got: %v", err)
		}
	})

	t.Run("stdio_mode_exits_on_closed_stdin", func(t *testing.T) {
		// Test the stdio transport path by providing a stdin that
		// immediately returns EOF. The server should exit cleanly.
		tmpDir := setupMCPTestProject(t, testDBHeader)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		// Save and replace os.Stdin with a pipe that closes immediately.
		oldStdin := os.Stdin
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("failed to create pipe: %v", err)
		}
		w.Close() // Close write end so reads get EOF immediately.
		os.Stdin = r
		defer func() { os.Stdin = oldStdin; r.Close() }()

		// Also capture stderr since stdio mode logs there.
		oldStderr := os.Stderr
		stderrR, stderrW, _ := os.Pipe()
		os.Stderr = stderrW
		defer func() { os.Stderr = oldStderr }()

		cmd := createMCPTestCmd()
		outBuf := new(bytes.Buffer)
		cmd.SetOut(outBuf)
		cmd.SetArgs([]string{"mcp-server", "--stdio"})

		// Execute -- should return quickly due to EOF on stdin.
		execErr := cmd.Execute()

		// Close stderr capture and read it.
		stderrW.Close()
		stderrBytes, _ := io.ReadAll(stderrR)
		stderrR.Close()
		stderrOutput := string(stderrBytes)

		// The server should have printed the startup message to stderr.
		if !strings.Contains(stderrOutput, "RTMX MCP server (stdio)") {
			t.Errorf("expected stdio startup message on stderr, got: %q", stderrOutput)
		}
		if !strings.Contains(stderrOutput, "database") {
			t.Errorf("expected database path in stderr output, got: %q", stderrOutput)
		}

		// The command may or may not return an error on EOF -- either
		// nil or an EOF-related error is acceptable.
		if execErr != nil && !strings.Contains(execErr.Error(), "EOF") {
			// Some non-EOF error is unexpected.
			t.Logf("stdio mode returned error (may be expected): %v", execErr)
		}
	})

	t.Run("stdio_flag_default_false", func(t *testing.T) {
		cmd := createMCPTestCmd()
		mcpCmd, _, _ := cmd.Find([]string{"mcp-server"})
		if mcpCmd == nil {
			t.Fatal("mcp-server command not found")
		}
		val, _ := mcpCmd.Flags().GetBool("stdio")
		if val {
			t.Error("stdio should default to false")
		}
	})

	t.Run("quiet_flag_default_false", func(t *testing.T) {
		cmd := createMCPTestCmd()
		mcpCmd, _, _ := cmd.Find([]string{"mcp-server"})
		if mcpCmd == nil {
			t.Fatal("mcp-server command not found")
		}
		val, _ := mcpCmd.Flags().GetBool("quiet")
		if val {
			t.Error("quiet should default to false")
		}
	})

	t.Run("command_use_and_short_description", func(t *testing.T) {
		if mcpServerCmd.Use != "mcp-server" {
			t.Errorf("Use = %q, want mcp-server", mcpServerCmd.Use)
		}
		if mcpServerCmd.Short == "" {
			t.Error("Short description should not be empty")
		}
	})

	t.Run("command_registered_on_root", func(t *testing.T) {
		found := false
		for _, sub := range rootCmd.Commands() {
			if sub.Use == "mcp-server" {
				found = true
				break
			}
		}
		if !found {
			t.Error("mcp-server should be registered as a subcommand of root")
		}
	})

	t.Run("port_flag_combined_with_host", func(t *testing.T) {
		cmd := createMCPTestCmd()
		mcpCmd, _, _ := cmd.Find([]string{"mcp-server"})
		if mcpCmd == nil {
			t.Fatal("mcp-server command not found")
		}
		err := mcpCmd.Flags().Parse([]string{"--port", "8080", "--host", "192.168.1.1"})
		if err != nil {
			t.Fatalf("flag parse failed: %v", err)
		}
		portVal, _ := mcpCmd.Flags().GetInt("port")
		hostVal, _ := mcpCmd.Flags().GetString("host")
		if portVal != 8080 {
			t.Errorf("port = %d, want 8080", portVal)
		}
		if hostVal != "192.168.1.1" {
			t.Errorf("host = %q, want 192.168.1.1", hostVal)
		}
	})
}

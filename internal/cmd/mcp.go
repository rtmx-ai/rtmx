package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rtmx-ai/rtmx/internal/adapters/mcp"
	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/spf13/cobra"
)

var (
	mcpPort  int
	mcpHost  string
	mcpStdio bool
	mcpQuiet bool
)

var mcpServerCmd = &cobra.Command{
	Use:   "mcp-server",
	Short: "Start MCP server for AI agent integration",
	Long: `Start a Model Context Protocol (MCP) server that exposes RTMX
operations as tools that AI agents can call.

Supports two transports:
  --stdio  JSON-RPC over stdin/stdout (default for Claude Code, Cursor)
  HTTP     JSON-RPC over HTTP on /mcp endpoint (default without --stdio)

Tools exposed:
  status   - RTM completion status
  backlog  - Prioritized incomplete requirements
  health   - Health checks on the RTM database
  deps     - Dependency information
  verify   - Verification status
  markers  - Test marker coverage
  next     - Highest-priority unblocked requirement
  claim    - Claim a requirement (multi-agent coordination)
  release  - Release a claimed requirement
  release_assign - Assign requirements to a version

Examples:
  rtmx mcp-server --stdio                     # stdio transport (for MCP clients)
  rtmx mcp-server                             # HTTP on default port 3000
  rtmx mcp-server --port 8080                 # Custom port
  rtmx mcp-server --host 0.0.0.0 --port 3000  # Bind to all interfaces

Claude Code setup:
  claude mcp add rtmx -- rtmx mcp-server --stdio

Cursor setup (.cursor/mcp.json):
  {"mcpServers":{"rtmx":{"command":"rtmx","args":["mcp-server","--stdio"]}}}`,
	RunE: runMCPServer,
}

func init() {
	mcpServerCmd.Flags().IntVar(&mcpPort, "port", 0, "port to listen on (default: from config or 3000)")
	mcpServerCmd.Flags().StringVar(&mcpHost, "host", "", "host to bind to (default: from config or localhost)")
	mcpServerCmd.Flags().BoolVar(&mcpStdio, "stdio", false, "use stdin/stdout transport (for Claude Code, Cursor)")
	mcpServerCmd.Flags().BoolVar(&mcpQuiet, "quiet", false, "suppress response size logging to stderr")

	rootCmd.AddCommand(mcpServerCmd)
}

func runMCPServer(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	dbPath := cfg.DatabasePath(cwd)

	// Determine host and port: flag > config > default
	host := mcpHost
	if host == "" {
		host = cfg.RTMX.MCP.Host
	}
	if host == "" {
		host = "localhost"
	}

	port := mcpPort
	if port == 0 {
		port = cfg.RTMX.MCP.Port
	}
	if port == 0 {
		port = 3000
	}

	srv := mcp.NewServer(dbPath, cfg, mcp.WithHost(host), mcp.WithPort(port), mcp.WithQuiet(mcpQuiet))

	if mcpStdio {
		// stdio transport: JSON-RPC over stdin/stdout
		// Log to stderr so stdout stays clean for the protocol
		_, _ = fmt.Fprintf(os.Stderr, "RTMX MCP server (stdio) -- database: %s\n", dbPath)
		return srv.StartStdio(os.Stdin, os.Stdout)
	}

	// HTTP transport
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cmd.Println("\nShutting down MCP server...")
		_ = srv.Shutdown(context.Background())
	}()

	cmd.Printf("RTMX MCP server listening on %s:%d\n", host, port)
	cmd.Printf("Database: %s\n", dbPath)
	cmd.Println("Press Ctrl+C to stop.")

	if err := srv.Start(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

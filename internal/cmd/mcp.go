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
	mcpPort int
	mcpHost string
)

var mcpServerCmd = &cobra.Command{
	Use:   "mcp-server",
	Short: "Start MCP server for AI agent integration",
	Long: `Start a Model Context Protocol (MCP) server that exposes RTMX
operations as tools that AI agents can call.

The server uses JSON-RPC 2.0 over HTTP and exposes the following tools:
  status   - RTM completion status
  backlog  - Prioritized incomplete requirements
  health   - Health checks on the RTM database
  deps     - Dependency information
  verify   - Verification status
  markers  - Test marker coverage

Examples:
  rtmx mcp-server                          # Start on default port 3000
  rtmx mcp-server --port 8080              # Custom port
  rtmx mcp-server --host 0.0.0.0 --port 3000  # Bind to all interfaces`,
	RunE: runMCPServer,
}

func init() {
	mcpServerCmd.Flags().IntVar(&mcpPort, "port", 0, "port to listen on (default: from config or 3000)")
	mcpServerCmd.Flags().StringVar(&mcpHost, "host", "", "host to bind to (default: from config or localhost)")

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

	srv := mcp.NewServer(dbPath, cfg, mcp.WithHost(host), mcp.WithPort(port))

	// Handle graceful shutdown
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

// Package agentpack loads the shared marketplace MCP pack (REQ-MONO-062b/j).
package agentpack

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed mcp.json
var embeddedMCP []byte

// Server is one MCP stdio server entry from the pack.
type Server struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type mcpFile struct {
	MCPServers map[string]Server `json:"mcpServers"`
}

// LoadMCP returns the canonical rtmx MCP server entry.
// Search order: RTMX_AGENT_PACK, walk-up plugins/rtmx/mcp.json, embedded copy.
func LoadMCP() (Server, string, error) {
	candidates := []string{}
	if pack := os.Getenv("RTMX_AGENT_PACK"); pack != "" {
		candidates = append(candidates, filepath.Join(pack, "mcp.json"))
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, walkPackMCP(cwd)...)
	}
	for _, path := range candidates {
		if path == "" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		srv, err := parseMCP(data)
		if err != nil {
			return Server{}, "", fmt.Errorf("parse %s: %w", path, err)
		}
		return srv, path, nil
	}
	srv, err := parseMCP(embeddedMCP)
	if err != nil {
		return Server{}, "", fmt.Errorf("parse embedded mcp.json: %w", err)
	}
	return srv, "embedded:internal/agentpack/mcp.json", nil
}

func walkPackMCP(start string) []string {
	var out []string
	dir := start
	for {
		out = append(out, filepath.Join(dir, "plugins", "rtmx", "mcp.json"))
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return out
}

func parseMCP(data []byte) (Server, error) {
	var f mcpFile
	if err := json.Unmarshal(data, &f); err != nil {
		return Server{}, err
	}
	srv, ok := f.MCPServers["rtmx"]
	if !ok {
		return Server{}, fmt.Errorf("mcpServers.rtmx missing")
	}
	if srv.Command == "" || len(srv.Args) == 0 {
		return Server{}, fmt.Errorf("mcpServers.rtmx command/args required")
	}
	return srv, nil
}

// EmbeddedMCP returns the embedded pack bytes (for parity tests).
func EmbeddedMCP() []byte {
	return append([]byte(nil), embeddedMCP...)
}

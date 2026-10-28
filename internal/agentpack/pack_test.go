package agentpack

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestEmbeddedMatchesPluginsPack(t *testing.T) {
	// Keep embed in lockstep with plugins/rtmx/mcp.json (REQ-MONO-062j / 062q).
	root := filepath.Join("..", "..", "plugins", "rtmx", "mcp.json")
	want, err := os.ReadFile(root)
	if err != nil {
		t.Fatalf("read pack: %v", err)
	}
	if !bytes.Equal(bytes.TrimSpace(want), bytes.TrimSpace(EmbeddedMCP())) {
		t.Fatalf("internal/agentpack/mcp.json diverged from plugins/rtmx/mcp.json")
	}
}

func TestLoadMCPFromEnvPack(t *testing.T) {
	dir := t.TempDir()
	pack := filepath.Join(dir, "mcp.json")
	body := []byte(`{"mcpServers":{"rtmx":{"command":"rtmx","args":["mcp-server","--stdio"]}}}`)
	if err := os.WriteFile(pack, body, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RTMX_AGENT_PACK", dir)
	srv, src, err := LoadMCP()
	if err != nil {
		t.Fatal(err)
	}
	if srv.Command != "rtmx" || len(srv.Args) != 2 || srv.Args[0] != "mcp-server" {
		t.Fatalf("unexpected server %#v from %s", srv, src)
	}
}

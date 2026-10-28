package agentpack

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWrapperMCPMatchesPack(t *testing.T) {
	root := filepath.Join("..", "..")
	packPath := filepath.Join(root, "plugins", "rtmx", "mcp.json")
	packData, err := os.ReadFile(packPath)
	if err != nil {
		t.Fatal(err)
	}
	var pack mcpFile
	if err := json.Unmarshal(packData, &pack); err != nil {
		t.Fatal(err)
	}
	want := pack.MCPServers["rtmx"]

	wrappers := []string{
		filepath.Join(root, "plugins", "claude-rtmx", ".mcp.json"),
		filepath.Join(root, "plugins", "cursor-rtmx", "mcp.json"),
	}
	for _, path := range wrappers {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		var f mcpFile
		if err := json.Unmarshal(data, &f); err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		got := f.MCPServers["rtmx"]
		if got.Command != want.Command || len(got.Args) != len(want.Args) {
			t.Fatalf("%s diverged from pack: got %#v want %#v", path, got, want)
		}
		for i := range want.Args {
			if got.Args[i] != want.Args[i] {
				t.Fatalf("%s args diverged: %#v vs %#v", path, got.Args, want.Args)
			}
		}
	}
}

func TestPluginManifestsHaveRequiredFields(t *testing.T) {
	root := filepath.Join("..", "..")
	type manifest struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
	}
	paths := []string{
		filepath.Join(root, "plugins", "claude-rtmx", ".claude-plugin", "plugin.json"),
		filepath.Join(root, "plugins", "cursor-rtmx", ".cursor-plugin", "plugin.json"),
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var m manifest
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if m.Name == "" || m.Version == "" || m.Description == "" {
			t.Fatalf("%s missing name/version/description: %#v", path, m)
		}
	}
}

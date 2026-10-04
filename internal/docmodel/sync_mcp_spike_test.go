package docmodel

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func schemasDir(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, "docs", "schemas")
}

func TestSyncMCPImpactSpike(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-001e")
	dir := schemasDir(t)
	spike, err := os.ReadFile(filepath.Join(dir, "SYNC_MCP_IMPACT_SPIKE.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(spike)
	for _, need := range []string{
		"requirement_fields",
		"Gap analysis",
		"Proposal",
		"MCP projection",
		"No breaking sync-protocol-v1",
		"managed-sync",
		"Sketch only",
	} {
		if !strings.Contains(text, need) {
			t.Errorf("SYNC_MCP_IMPACT_SPIKE.md missing %q", need)
		}
	}

	raw, err := os.ReadFile(filepath.Join(dir, "sync-protocol-v1.1-ac-extension.stub.json"))
	if err != nil {
		t.Fatal(err)
	}
	var stub map[string]any
	if err := json.Unmarshal(raw, &stub); err != nil {
		t.Fatal(err)
	}
	compat, _ := stub["compatibility"].(map[string]any)
	if compat["breaking_change_required"] != false {
		t.Fatal("stub must state breaking_change_required=false for managed-sync launch")
	}
	if _, ok := stub["additive_shared_types"]; !ok {
		t.Fatal("stub must sketch additive_shared_types")
	}
}

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/markers"
)

func TestMarkersCheckInvocableAsCLIOrLibrary(t *testing.T) {
	dir := t.TempDir()
	rtmxDir := filepath.Join(dir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rtmx.yaml"), []byte("rtmx:\n  database: .rtmx/database.csv\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	csv := "req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id\nREQ-AUTH-001,CLI,Feature,ok,,,,,,COMPLETE,HIGH,1,,,,,,,,,,\n"
	if err := os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "ok.go")
	if err := os.WriteFile(src, []byte("// @req REQ-AUTH-001\npackage ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Library path
	issues := markers.CheckSources([]string{src}, map[string]struct{}{"REQ-AUTH-001": {}})
	if markers.ExitCode(issues) != 0 {
		t.Fatalf("library check failed: %#v", issues)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	markersCheckRoot = dir
	if err := runMarkersCheck(markersCheckCmd, nil); err != nil {
		t.Fatalf("CLI check failed: %v", err)
	}
}

func TestMarkersCheckDeliveryHasNoIDEOrForgePlugin(t *testing.T) {
	root := filepath.Join("..", "..")
	plugins := filepath.Join(root, "plugins")
	err := filepath.Walk(plugins, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		name := strings.ToLower(info.Name())
		base := strings.ToLower(path)
		if strings.Contains(name, "marker") && (strings.Contains(base, "vscode") ||
			strings.Contains(base, "jetbrains") ||
			strings.Contains(base, "forge") ||
			strings.Contains(base, "gitea") ||
			strings.Contains(base, "github-app")) {
			t.Errorf("marker checker must not ship as IDE/forge plugin: %s", path)
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	// Delivery locations are the library and CLI only.
	if _, err := os.Stat(filepath.Join(root, "internal", "markers", "check.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "internal", "cmd", "markers_check.go")); err != nil {
		t.Fatal(err)
	}
}


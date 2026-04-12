package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestMarkersDiscover(t *testing.T) {
	rtmx.Req(t, "REQ-GO-034")

	dbContent := `req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id
REQ-001,CLI,Feature,First requirement,Works,test_file.go,TestFirst,Unit Test,COMPLETE,HIGH,1,,,,,,,,,
REQ-002,CLI,Feature,Second requirement,Works,test_file.go,TestSecond,Unit Test,MISSING,HIGH,1,,,,,,,,,
REQ-003,CLI,Feature,Third requirement no marker,Works,,,Unit Test,MISSING,HIGH,1,,,,,,,,,
`

	t.Run("list_markers", func(t *testing.T) {
		tmpDir := t.TempDir()
		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		_ = os.MkdirAll(rtmxDir, 0755)
		_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)

		testDir := filepath.Join(tmpDir, "internal", "cmd")
		_ = os.MkdirAll(testDir, 0755)
		goTest := "package cmd\n\nimport (\n\t\"testing\"\n\t\"github.com/rtmx-ai/rtmx/pkg/rtmx\"\n)\n\nfunc TestFirst(t *testing.T) {\n\trtmx.Req(t, \"REQ-001\")\n}\n\nfunc TestSecond(t *testing.T) {\n\trtmx.Req(t, \"REQ-002\")\n}\n"
		_ = os.WriteFile(filepath.Join(testDir, "feature_test.go"), []byte(goTest), 0644)

		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		buf := new(bytes.Buffer)
		markersCmd.SetOut(buf)
		markersShowMissing = false
		err := markersCmd.RunE(markersCmd, nil)
		if err != nil {
			t.Fatalf("markers command failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "REQ-001") {
			t.Errorf("expected REQ-001 in output, got:\n%s", out)
		}
		if !strings.Contains(out, "REQ-002") {
			t.Errorf("expected REQ-002 in output, got:\n%s", out)
		}
		if !strings.Contains(out, "markers found") {
			t.Errorf("expected summary in output, got:\n%s", out)
		}
	})

	t.Run("show_missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		_ = os.MkdirAll(rtmxDir, 0755)
		_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)

		testDir := filepath.Join(tmpDir, "internal", "cmd")
		_ = os.MkdirAll(testDir, 0755)
		goTest := "package cmd\n\nimport (\n\t\"testing\"\n\t\"github.com/rtmx-ai/rtmx/pkg/rtmx\"\n)\n\nfunc TestFirst(t *testing.T) {\n\trtmx.Req(t, \"REQ-001\")\n}\n"
		_ = os.WriteFile(filepath.Join(testDir, "feature_test.go"), []byte(goTest), 0644)

		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		buf := new(bytes.Buffer)
		markersCmd.SetOut(buf)
		markersShowMissing = true
		err := markersCmd.RunE(markersCmd, nil)
		if err != nil {
			t.Fatalf("markers --missing failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "REQ-003") {
			t.Errorf("expected REQ-003 (no marker) in missing output, got:\n%s", out)
		}
		if !strings.Contains(out, "without markers") {
			t.Errorf("expected 'without markers' summary, got:\n%s", out)
		}
	})
}

func TestExtractGoCommentMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-001")

	goTest := `package example

import "testing"

func TestAuth(t *testing.T) {
	// rtmx:req REQ-AUTH-001
	t.Log("test auth")
}

func TestLogin(t *testing.T) {
	// rtmx:req REQ-AUTH-002
	t.Log("test login")
}

func TestNoMarker(t *testing.T) {
	t.Log("no marker here")
}
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "auth_test.go")
	_ = os.WriteFile(testFile, []byte(goTest), 0644)

	markers, err := extractGoMarkersFromFile(testFile)
	if err != nil {
		t.Fatalf("extractGoMarkersFromFile() error = %v", err)
	}
	if len(markers) != 2 {
		t.Fatalf("expected 2 markers, got %d", len(markers))
	}
	if markers[0].ReqID != "REQ-AUTH-001" {
		t.Errorf("markers[0].ReqID = %q, want REQ-AUTH-001", markers[0].ReqID)
	}
	if markers[0].TestFunction != "TestAuth" {
		t.Errorf("markers[0].TestFunction = %q, want TestAuth", markers[0].TestFunction)
	}
	if markers[1].ReqID != "REQ-AUTH-002" {
		t.Errorf("markers[1].ReqID = %q, want REQ-AUTH-002", markers[1].ReqID)
	}
	if markers[1].TestFunction != "TestLogin" {
		t.Errorf("markers[1].TestFunction = %q, want TestLogin", markers[1].TestFunction)
	}
}

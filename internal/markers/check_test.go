package markers_test

import (
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/markers"
)

func TestCheckerExitsNonZeroForDanglingMarker(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-068b")

	dir := t.TempDir()
	src := filepath.Join(dir, "sample.go")
	if err := os.WriteFile(src, []byte("// @req REQ-AUTH-999\npackage sample\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	known := map[string]struct{}{"REQ-AUTH-001": {}}
	issues := markers.CheckSources([]string{src}, known)
	if markers.ExitCode(issues) == 0 {
		t.Fatal("expected non-zero exit for dangling marker")
	}
	found := false
	for _, iss := range issues {
		if iss.Kind == "dangling" && iss.Token == "REQ-AUTH-999" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected dangling REQ-AUTH-999, got %#v", issues)
	}
}

func TestCheckerExitsNonZeroForMalformedMarker(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "sample.py")
	if err := os.WriteFile(src, []byte("# REQ-bad-id\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	issues := markers.CheckSources([]string{src}, map[string]struct{}{})
	if markers.ExitCode(issues) == 0 {
		t.Fatal("expected non-zero exit for malformed marker")
	}
	found := false
	for _, iss := range issues {
		if iss.Kind == "malformed" && iss.Token == "REQ-bad-id" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected malformed REQ-bad-id, got %#v", issues)
	}
}


package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestMigrateDocumentCommand(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-003")
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	input := filepath.Join(root, "internal", "docmodel", "migrate", "testdata", "fixture")
	out := t.TempDir()

	migrateDocInput = input
	migrateDocOutput = out
	migrateDocPrevious = ""
	defer func() {
		migrateDocInput = ""
		migrateDocOutput = ""
		migrateDocPrevious = ""
	}()

	buf := new(bytes.Buffer)
	cmd := migrateDocumentCmd
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := runMigrateDocument(cmd, nil); err != nil {
		t.Fatal(err)
	}
	outText := buf.String()
	if !strings.Contains(outText, "unchanged") {
		t.Fatalf("expected no default flip notice:\n%s", outText)
	}
	for _, name := range []string{"requirements.jsonl", "requirements.document.json", "database.projection.csv"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}

	// Second pass with --previous keeps stable ids
	migrateDocPrevious = filepath.Join(out, "requirements.jsonl")
	out2 := t.TempDir()
	migrateDocOutput = out2
	if err := runMigrateDocument(cmd, nil); err != nil {
		t.Fatal(err)
	}
}

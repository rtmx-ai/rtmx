package cmd

import (
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/proofbundle"
	"github.com/spf13/cobra"
)

func TestProofExportWritesLocalBundle(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-065a")

	dir := t.TempDir()
	rtmxDir := filepath.Join(dir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rtmx.yaml"), []byte("rtmx:\n  database: .rtmx/database.csv\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte("req_id,status\nREQ-MONO-065ab,COMPLETE\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sha := "cccccccccccccccccccccccccccccccccccccccc"
	line := `{"requirement_id":"REQ-MONO-065ab","test_id":"t","commit_sha":"` + sha + `","result":"pass","timestamp":"2026-11-06T12:00:00Z"}` + "\n"
	if err := os.WriteFile(filepath.Join(rtmxDir, "evidence.jsonl"), []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join(dir, "out-bundle")
	proofExportOut = outDir
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := runProofExport(cmd, nil); err != nil {
		t.Fatal(err)
	}
	matrix, err := os.ReadFile(filepath.Join(outDir, proofbundle.MatrixName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(matrix), "REQ-MONO-065ab") {
		t.Fatalf("matrix missing requirement: %s", matrix)
	}
	evidence, err := os.ReadFile(filepath.Join(outDir, proofbundle.EvidenceName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(evidence), sha) {
		t.Fatalf("evidence missing sha: %s", evidence)
	}
}

func TestProofExportOpensNoNetworkConnection(t *testing.T) {
	dir := t.TempDir()
	rtmxDir := filepath.Join(dir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rtmx.yaml"), []byte("rtmx:\n  database: .rtmx/database.csv\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte("req_id,status\nREQ-MONO-065ab,COMPLETE\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rtmxDir, "evidence.jsonl"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	// Outbound network is unavailable: export must finish without dialing.
	proofExportOut = filepath.Join(dir, "airgap-bundle")
	cmd := &cobra.Command{}
	if err := runProofExport(cmd, nil); err != nil {
		t.Fatal(err)
	}
	if n := ProofOutboundCalls(); n != 0 {
		t.Fatalf("outbound dials = %d, want 0", n)
	}
	if _, err := os.Stat(filepath.Join(proofExportOut, proofbundle.MatrixName)); err != nil {
		t.Fatal(err)
	}
}


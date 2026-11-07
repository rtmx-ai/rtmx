package proofbundle_test

import (
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/proofbundle"
)

const (
	oldSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	newSHA = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func TestBundleContainsMatrixEvidenceAndCommitSHAs(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-065a")

	dir := t.TempDir()
	matrix := filepath.Join(dir, "rtm_database.csv")
	evidence := filepath.Join(dir, "evidence.jsonl")
	if err := os.WriteFile(matrix, []byte("req_id,status\nREQ-MONO-065aa,COMPLETE\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lines := strings.Join([]string{
		`{"requirement_id":"REQ-MONO-065aa","test_id":"t","commit_sha":"` + oldSHA + `","result":"pass","timestamp":"2026-11-06T10:00:00Z"}`,
		`{"requirement_id":"REQ-MONO-065aa","test_id":"t","commit_sha":"` + newSHA + `","result":"pass","timestamp":"2026-11-06T11:00:00Z"}`,
		"",
	}, "\n")
	if err := os.WriteFile(evidence, []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "bundle")
	if err := proofbundle.Build(out, matrix, evidence); err != nil {
		t.Fatal(err)
	}

	gotMatrix, err := os.ReadFile(filepath.Join(out, proofbundle.MatrixName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gotMatrix), "REQ-MONO-065aa") {
		t.Fatalf("matrix missing requirement: %s", gotMatrix)
	}
	gotEvidence, err := os.ReadFile(filepath.Join(out, proofbundle.EvidenceName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gotEvidence), oldSHA) || !strings.Contains(string(gotEvidence), newSHA) {
		t.Fatalf("evidence missing SHAs: %s", gotEvidence)
	}
	raw, err := os.ReadFile(filepath.Join(out, proofbundle.SHAsName))
	if err != nil {
		t.Fatal(err)
	}
	var shas []string
	if err := json.Unmarshal(raw, &shas); err != nil {
		t.Fatal(err)
	}
	if len(shas) != 2 || shas[0] != oldSHA || shas[1] != newSHA {
		t.Fatalf("commit_shas.json = %#v", shas)
	}
}

func TestBundleDoesNotContainGitBlobsTreesOrPatches(t *testing.T) {
	dir := t.TempDir()
	matrix := filepath.Join(dir, "rtm_database.csv")
	evidence := filepath.Join(dir, "evidence.jsonl")
	if err := os.WriteFile(matrix, []byte("req_id,status\nREQ-MONO-065aa,COMPLETE\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	line := `{"requirement_id":"REQ-MONO-065aa","test_id":"t","commit_sha":"` + oldSHA + `","result":"pass","timestamp":"2026-11-06T10:00:00Z","blob":"deadbeef","tree":"cafebabe","patch":"@@ -1 +1 @@","content":"TOP SECRET"}` + "\n"
	if err := os.WriteFile(evidence, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "bundle")
	if err := proofbundle.Build(out, matrix, evidence); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]struct{}{}
	for _, e := range entries {
		names[e.Name()] = struct{}{}
	}
	for _, forbidden := range []string{"blob", "tree", "patch", "objects", "pack"} {
		if _, ok := names[forbidden]; ok {
			t.Fatalf("bundle contains git path %q", forbidden)
		}
	}
	gotEvidence, err := os.ReadFile(filepath.Join(out, proofbundle.EvidenceName))
	if err != nil {
		t.Fatal(err)
	}
	body := string(gotEvidence)
	for _, needle := range []string{"deadbeef", "cafebabe", "@@ -1 +1 @@", "TOP SECRET", `"blob"`, `"tree"`, `"patch"`, `"content"`} {
		if strings.Contains(body, needle) {
			t.Fatalf("evidence still contains %q: %s", needle, body)
		}
	}
	var rec map[string]any
	if err := json.Unmarshal(gotEvidence, &rec); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"blob", "tree", "patch", "content"} {
		if _, ok := rec[k]; ok {
			t.Fatalf("evidence key %q present: %#v", k, rec)
		}
	}
}


func TestBundleContainsHashManifest(t *testing.T) {
	dir := t.TempDir()
	matrix := filepath.Join(dir, "rtm_database.csv")
	evidence := filepath.Join(dir, "evidence.jsonl")
	if err := os.WriteFile(matrix, []byte("req_id,status\nREQ-MONO-065b,COMPLETE\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(evidence, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "bundle")
	if err := proofbundle.Build(out, matrix, evidence); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(out, proofbundle.ManifestName))
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	for _, name := range []string{proofbundle.MatrixName, proofbundle.EvidenceName, proofbundle.SHAsName} {
		if !strings.Contains(body, name) {
			t.Fatalf("manifest missing %s: %s", name, body)
		}
	}
}

func TestManifestVerifySucceedsOffline(t *testing.T) {
	dir := t.TempDir()
	matrix := filepath.Join(dir, "rtm_database.csv")
	evidence := filepath.Join(dir, "evidence.jsonl")
	if err := os.WriteFile(matrix, []byte("req_id,status\nREQ-MONO-065b,COMPLETE\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(evidence, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "bundle")
	if err := proofbundle.Build(out, matrix, evidence); err != nil {
		t.Fatal(err)
	}
	// Outbound network unavailable: verification is local file hashing only.
	if err := proofbundle.Verify(out); err != nil {
		t.Fatal(err)
	}
}

func TestChangedByteFailsManifestVerification(t *testing.T) {
	dir := t.TempDir()
	matrix := filepath.Join(dir, "rtm_database.csv")
	evidence := filepath.Join(dir, "evidence.jsonl")
	if err := os.WriteFile(matrix, []byte("req_id,status\nREQ-MONO-065b,COMPLETE\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(evidence, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "bundle")
	if err := proofbundle.Build(out, matrix, evidence); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(out, proofbundle.MatrixName)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	tampered := append([]byte{}, raw...)
	tampered[0] ^= 0xff
	if err := os.WriteFile(path, tampered, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := proofbundle.Verify(out); err == nil {
		t.Fatal("expected verification failure after changing a bundled byte")
	}
}

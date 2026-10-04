package acverify

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestEvaluateDocumentACMatrix(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-002")
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..")) // acverify → docmodel → internal → repo
	docPath := filepath.Join(root, "docs", "schemas", "fixtures", "requirement-document-v0.valid.json")

	doc, err := LoadDocumentFile(docPath)
	if err != nil {
		t.Fatal(err)
	}

	// No evidence → gaps / MISSING or PARTIAL depending on required ACs
	none := EvaluateDocument(doc, nil)
	if len(none) < 1 {
		t.Fatal("expected results")
	}
	foundAuth := false
	for _, r := range none {
		if r.ReqID == "REQ-AUTH-001" {
			foundAuth = true
			if r.PolicyName != PolicyATDDRequiredAll {
				t.Fatalf("policy = %s", r.PolicyName)
			}
			if r.Status == "COMPLETE" {
				t.Fatal("no evidence must not yield COMPLETE")
			}
			out := FormatMatrix(r)
			if !strings.Contains(out, "AC matrix:") {
				t.Fatalf("missing matrix:\n%s", out)
			}
		}
	}
	if !foundAuth {
		t.Fatal("expected REQ-AUTH-001 in document")
	}

	// Evidence for both AUTH ACs → COMPLETE
	evidence := []EvidenceHit{
		{TestName: "TestRoomRejectsMissingToken", ReqID: "REQ-AUTH-001", Passed: true},
		{TestName: "TestAuthStatusWithValidToken", ReqID: "REQ-AUTH-001", Passed: true},
	}
	with := EvaluateDocument(doc, evidence)
	for _, r := range with {
		if r.ReqID != "REQ-AUTH-001" {
			continue
		}
		if r.Status != "COMPLETE" {
			t.Fatalf("status = %s want COMPLETE\n%s", r.Status, FormatMatrix(r))
		}
	}
}

func TestLoadDocumentFileRejectsBadVersion(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-002")
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":"nope","requirements":[{"req_id":"REQ-X-1","acs":[{"ac_id":"AC-1","statement":"s"}],"status":"MISSING","priority":"HIGH","requirement_text":"t"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadDocumentFile(path); err == nil {
		t.Fatal("expected schema_version error")
	}
}

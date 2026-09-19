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

func TestVerifyACDocumentReportOnly(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-002")
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	doc := filepath.Join(root, "docs", "schemas", "fixtures", "requirement-document-v0.valid.json")

	verifyACDocument = doc
	verifyResults = ""
	verifyUpdate = false
	defer func() {
		verifyACDocument = ""
		verifyResults = ""
		verifyUpdate = false
	}()

	buf := new(bytes.Buffer)
	cmd := verifyCmd
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	err := runVerifyACDocument(cmd)
	out := buf.String()
	if !strings.Contains(out, "AC matrix:") {
		t.Fatalf("expected AC matrix output:\n%s", out)
	}
	if !strings.Contains(out, "report-only") {
		t.Fatalf("expected report-only banner:\n%s", out)
	}
	if !strings.Contains(out, acverifyPolicyName()) {
		t.Fatalf("expected policy name in output:\n%s", out)
	}
	if err == nil {
		t.Fatal("expected non-COMPLETE without evidence")
	}

	// Refuse --update
	verifyUpdate = true
	if err := runVerify(cmd, nil); err == nil || !strings.Contains(err.Error(), "report-only") {
		t.Fatalf("expected refuse --update, got %v", err)
	}
	verifyUpdate = false

	// With full AUTH evidence still incomplete overall (other reqs gap) — write results for AUTH only
	resultsPath := filepath.Join(t.TempDir(), "results.json")
	payload := `[
  {"marker":{"req_id":"REQ-AUTH-001","test_name":"TestRoomRejectsMissingToken","test_file":"t.go"},"passed":true},
  {"marker":{"req_id":"REQ-AUTH-001","test_name":"TestAuthStatusWithValidToken","test_file":"t.go"},"passed":true}
]`
	if err := os.WriteFile(resultsPath, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	verifyResults = resultsPath
	buf.Reset()
	err = runVerifyACDocument(cmd)
	out = buf.String()
	if !strings.Contains(out, "REQ-AUTH-001") || !strings.Contains(out, "COMPLETE") {
		t.Fatalf("expected AUTH COMPLETE:\n%s", out)
	}
	if err == nil {
		t.Fatal("other requirements still incomplete; want error")
	}
}

func acverifyPolicyName() string {
	return "atdd-required-all-v0"
}

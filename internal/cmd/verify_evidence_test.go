package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestVerifyEvidenceRequiresResolvableSHA(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-064b")

	t.Run("no commit writes nothing", func(t *testing.T) {
		dir := t.TempDir()
		git := exec.Command("git", "init")
		git.Dir = dir
		if out, err := git.CombinedOutput(); err != nil {
			t.Fatalf("git init: %v\n%s", err, out)
		}
		db := database.NewDatabase()
		req := database.NewRequirement("REQ-MONO-064b")
		req.TestModule = "internal/cmd/verify_evidence_test.go"
		req.TestFunction = "TestVerifyEvidenceRequiresResolvableSHA"
		if err := db.Add(req); err != nil {
			t.Fatal(err)
		}
		n, err := recordVerifyEvidence(dir, db, []VerificationResult{{
			ReqID:       req.ReqID,
			TestsPassed: 1,
		}})
		if err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Fatalf("wrote %d records, want 0", n)
		}
		if _, err := os.Stat(filepath.Join(dir, ".rtmx", "evidence.jsonl")); !os.IsNotExist(err) {
			t.Fatalf("evidence file exists without a commit: %v", err)
		}
	})

	t.Run("resolvable HEAD writes one record", func(t *testing.T) {
		dir := t.TempDir()
		runGit(t, dir, "init")
		runGit(t, dir, "config", "user.email", "evidence@example.com")
		runGit(t, dir, "config", "user.name", "evidence")
		if err := os.WriteFile(filepath.Join(dir, "README"), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		runGit(t, dir, "add", "README")
		runGit(t, dir, "-c", "commit.gpgsign=false", "commit", "-m", "init")
		sha := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))

		db := database.NewDatabase()
		req := database.NewRequirement("REQ-MONO-064b")
		req.TestModule = "internal/cmd/verify_evidence_test.go"
		req.TestFunction = "TestVerifyEvidenceRequiresResolvableSHA"
		if err := db.Add(req); err != nil {
			t.Fatal(err)
		}
		n, err := recordVerifyEvidence(dir, db, []VerificationResult{{
			ReqID:       req.ReqID,
			TestsPassed: 1,
		}})
		if err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Fatalf("wrote %d records, want 1", n)
		}
		raw, err := os.ReadFile(filepath.Join(dir, ".rtmx", "evidence.jsonl"))
		if err != nil {
			t.Fatal(err)
		}
		var rec evidenceRecord
		if err := json.Unmarshal(raw, &rec); err != nil {
			t.Fatal(err)
		}
		if rec.CommitSHA != sha {
			t.Fatalf("sha = %s, want %s", rec.CommitSHA, sha)
		}
		if rec.RequirementID != req.ReqID {
			t.Fatalf("requirement = %s", rec.RequirementID)
		}
		if rec.TestID != req.TestModule+"::"+req.TestFunction {
			t.Fatalf("test = %s", rec.TestID)
		}
		if rec.Result != "pass" || rec.Timestamp == "" {
			t.Fatalf("record = %+v", rec)
		}
	})
}

func TestStructuralGateRejectsNonProvingResults(t *testing.T) {
	cases := []VerificationResult{
		{TestsSkipped: 1},
		{TestsPassed: 1, EmptyBody: true},
		{TestsPassed: 1, ZeroAssertions: true},
	}
	for _, c := range cases {
		if StructuralRejection(c) == "" {
			t.Fatalf("gate accepted %+v", c)
		}
	}
}

func TestRejectedNonProvingResultsAreNotPassEvidence(t *testing.T) {
	cases := []VerificationResult{
		{TestsSkipped: 1},
		{TestsPassed: 1, EmptyBody: true},
		{TestsPassed: 1, ZeroAssertions: true},
	}
	db := database.NewDatabase()
	req := database.NewRequirement("REQ-MONO-064e")
	req.TestModule = "internal/cmd/verify_evidence_test.go"
	req.TestFunction = "TestRejectedNonProvingResultsAreNotPassEvidence"
	if err := db.Add(req); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		c.ReqID = req.ReqID
		dir := evidenceRepo(t)
		n, err := recordVerifyEvidence(dir, db, []VerificationResult{c})
		if err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Fatalf("wrote %d records for %+v", n, c)
		}
		if raw, err := os.ReadFile(filepath.Join(dir, ".rtmx", "evidence.jsonl")); err == nil && strings.Contains(string(raw), `"result":"pass"`) {
			t.Fatalf("recorded pass for %+v: %s", c, raw)
		}
	}
}

func TestStructuralGateRejectsZeroAssertions(t *testing.T) {
	if got := StructuralRejection(VerificationResult{TestsPassed: 1, ZeroAssertions: true}); got != "zero-assertions" {
		t.Fatalf("rejection = %q, want zero-assertions", got)
	}
	if got := StructuralRejection(VerificationResult{TestsPassed: 1}); got != "" {
		t.Fatalf("asserting result rejected: %q", got)
	}
}

func TestRejectedZeroAssertionsIsNotPassEvidence(t *testing.T) {
	dir := evidenceRepo(t)
	db := database.NewDatabase()
	req := database.NewRequirement("REQ-MONO-064ec")
	req.TestModule = "internal/cmd/verify_evidence_test.go"
	req.TestFunction = "TestRejectedZeroAssertionsIsNotPassEvidence"
	if err := db.Add(req); err != nil {
		t.Fatal(err)
	}
	n, err := recordVerifyEvidence(dir, db, []VerificationResult{{
		ReqID:          req.ReqID,
		TestsPassed:    1,
		ZeroAssertions: true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("wrote %d records for zero assertions, want 0", n)
	}
	if raw, err := os.ReadFile(filepath.Join(dir, ".rtmx", "evidence.jsonl")); err == nil && strings.Contains(string(raw), `"result":"pass"`) {
		t.Fatalf("zero-assertion result recorded as pass: %s", raw)
	}
}

func TestStructuralGateRejectsEmptyBody(t *testing.T) {
	if got := StructuralRejection(VerificationResult{TestsPassed: 1, EmptyBody: true}); got != "empty-body" {
		t.Fatalf("rejection = %q, want empty-body", got)
	}
	if got := StructuralRejection(VerificationResult{TestsPassed: 1}); got != "" {
		t.Fatalf("non-empty body rejected: %q", got)
	}
}

func TestRejectedEmptyBodyIsNotPassEvidence(t *testing.T) {
	dir := evidenceRepo(t)
	db := database.NewDatabase()
	req := database.NewRequirement("REQ-MONO-064eb")
	req.TestModule = "internal/cmd/verify_evidence_test.go"
	req.TestFunction = "TestRejectedEmptyBodyIsNotPassEvidence"
	if err := db.Add(req); err != nil {
		t.Fatal(err)
	}
	n, err := recordVerifyEvidence(dir, db, []VerificationResult{{
		ReqID:       req.ReqID,
		TestsPassed: 1,
		EmptyBody:   true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("wrote %d records for an empty body, want 0", n)
	}
	if raw, err := os.ReadFile(filepath.Join(dir, ".rtmx", "evidence.jsonl")); err == nil && strings.Contains(string(raw), `"result":"pass"`) {
		t.Fatalf("empty body recorded as pass: %s", raw)
	}
}

func TestStructuralGateRejectsSkipped(t *testing.T) {
	if got := StructuralRejection(VerificationResult{TestsSkipped: 1}); got != "skipped" {
		t.Fatalf("rejection = %q, want skipped", got)
	}
	if got := StructuralRejection(VerificationResult{TestsPassed: 1}); got != "" {
		t.Fatalf("pass rejected: %q", got)
	}
	if got := StructuralRejection(VerificationResult{TestsFailed: 1, TestsSkipped: 1}); got != "" {
		t.Fatalf("failure rejected as a skip: %q", got)
	}
}

func TestRejectedSkipIsNotPassEvidence(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "evidence@example.com")
	runGit(t, dir, "config", "user.name", "evidence")
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "README")
	runGit(t, dir, "-c", "commit.gpgsign=false", "commit", "-m", "init")

	db := database.NewDatabase()
	req := database.NewRequirement("REQ-MONO-064ea")
	req.TestModule = "internal/cmd/verify_evidence_test.go"
	req.TestFunction = "TestRejectedSkipIsNotPassEvidence"
	if err := db.Add(req); err != nil {
		t.Fatal(err)
	}
	n, err := recordVerifyEvidence(dir, db, []VerificationResult{{
		ReqID:        req.ReqID,
		TestsSkipped: 1,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("wrote %d records for a skip, want 0", n)
	}
	raw, err := os.ReadFile(filepath.Join(dir, ".rtmx", "evidence.jsonl"))
	if err == nil && strings.Contains(string(raw), `"result":"pass"`) {
		t.Fatalf("skip recorded as pass: %s", raw)
	}
}

func TestVerifyEvidenceAppendOnlyAcrossSHAs(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-064c")

	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "evidence@example.com")
	runGit(t, dir, "config", "user.name", "evidence")
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "README")
	runGit(t, dir, "-c", "commit.gpgsign=false", "commit", "-m", "first")
	sha1 := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))

	db := database.NewDatabase()
	req := database.NewRequirement("REQ-MONO-064c")
	req.TestModule = "internal/cmd/verify_evidence_test.go"
	req.TestFunction = "TestVerifyEvidenceAppendOnlyAcrossSHAs"
	if err := db.Add(req); err != nil {
		t.Fatal(err)
	}
	if _, err := recordVerifyEvidence(dir, db, []VerificationResult{{
		ReqID:       req.ReqID,
		TestsPassed: 1,
	}}); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "README")
	runGit(t, dir, "-c", "commit.gpgsign=false", "commit", "-m", "second")
	sha2 := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))
	if sha1 == sha2 {
		t.Fatal("expected a new commit SHA")
	}
	if _, err := recordVerifyEvidence(dir, db, []VerificationResult{{
		ReqID:       req.ReqID,
		TestsPassed: 1,
	}}); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, ".rtmx", "evidence.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var shas []string
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var rec evidenceRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatal(err)
		}
		if rec.RequirementID != req.ReqID {
			t.Fatalf("requirement = %s", rec.RequirementID)
		}
		shas = append(shas, rec.CommitSHA)
	}
	if len(shas) != 2 || shas[0] != sha1 || shas[1] != sha2 {
		t.Fatalf("shas = %v, want [%s %s]", shas, sha1, sha2)
	}
}

func evidenceRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "evidence@example.com")
	runGit(t, dir, "config", "user.name", "evidence")
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "README")
	runGit(t, dir, "-c", "commit.gpgsign=false", "commit", "-m", "init")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

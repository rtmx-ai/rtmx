package proofbundle_test

import (
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"os"
	"strings"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/proofbundle"
)

func TestProjectionShowsVerifiedUnverifiedAndStaleCounts(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-065c")

	dir := t.TempDir()
	matrix := filepath.Join(dir, "matrix.csv")
	evidence := filepath.Join(dir, "evidence.jsonl")
	rev := "dddddddddddddddddddddddddddddddddddddddd"
	old := "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	if err := os.WriteFile(matrix, []byte("req_id,status\nREQ-A,COMPLETE\nREQ-B,NOT_STARTED\nREQ-C,PARTIAL\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lines := "" +
		`{"requirement_id":"REQ-A","test_id":"t","commit_sha":"` + rev + `","result":"pass","timestamp":"2026-11-06T10:00:00Z"}` + "\n" +
		`{"requirement_id":"REQ-C","test_id":"t","commit_sha":"` + old + `","result":"pass","timestamp":"2026-11-06T09:00:00Z"}` + "\n"
	if err := os.WriteFile(evidence, []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := proofbundle.ProjectCounts(matrix, evidence, rev)
	if err != nil {
		t.Fatal(err)
	}
	if got.Verified != 1 || got.Unverified != 1 || got.Stale != 1 {
		t.Fatalf("counts = %+v, want verified=1 unverified=1 stale=1", got)
	}
}

func TestStaleMeansLatestEvidenceSHADiffersFromRevisionUnderReview(t *testing.T) {
	dir := t.TempDir()
	matrix := filepath.Join(dir, "matrix.csv")
	evidence := filepath.Join(dir, "evidence.jsonl")
	rev := "ffffffffffffffffffffffffffffffffffffffff"
	older := "1111111111111111111111111111111111111111"
	newerOther := "2222222222222222222222222222222222222222"
	if err := os.WriteFile(matrix, []byte("req_id,status\nREQ-STALE,COMPLETE\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Latest line wins: pinned to newerOther, not the revision under review.
	lines := "" +
		`{"requirement_id":"REQ-STALE","test_id":"t","commit_sha":"` + older + `","result":"pass","timestamp":"2026-11-06T08:00:00Z"}` + "\n" +
		`{"requirement_id":"REQ-STALE","test_id":"t","commit_sha":"` + newerOther + `","result":"pass","timestamp":"2026-11-06T09:00:00Z"}` + "\n"
	if err := os.WriteFile(evidence, []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := proofbundle.ProjectCounts(matrix, evidence, rev)
	if err != nil {
		t.Fatal(err)
	}
	if got.Stale != 1 || got.Verified != 0 || got.Unverified != 0 {
		t.Fatalf("counts = %+v, want stale=1 only", got)
	}
}


func TestProjectionHasNoGanttSprintBoardOrPortfolioPlan(t *testing.T) {
	body, err := proofbundle.RenderProjection(proofbundle.Counts{
		Verified:   2,
		Unverified: 3,
		Stale:      1,
	})
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(body)
	for _, forbidden := range []string{"gantt", "sprint", "portfolio", "board", "chart"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("projection contains %q: %s", forbidden, body)
		}
	}
	if !strings.Contains(body, `"verified"`) || !strings.Contains(body, `"unverified"`) || !strings.Contains(body, `"stale"`) {
		t.Fatalf("projection missing count keys: %s", body)
	}
}

package prdextract_test

import (
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/prdextract"
)

func TestExtractorAcceptsPlainTextAndMarkdownPRD(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-067a")

	if err := prdextract.Accept(prdextract.Input{
		Format:  prdextract.FormatText,
		Content: "The system shall authenticate users.",
	}); err != nil {
		t.Fatalf("plain text: %v", err)
	}
	if err := prdextract.Accept(prdextract.Input{
		Format:  prdextract.FormatMarkdown,
		Content: "# PRD\n\nThe system shall authenticate users.\n",
	}); err != nil {
		t.Fatalf("markdown: %v", err)
	}
}

func TestEachCandidateHasStatementAndSourceSpan(t *testing.T) {
	content := "# PRD\n\nThe system shall authenticate users.\nThe system shall log access.\n"
	cands, err := prdextract.Extract(prdextract.Input{
		Format:  prdextract.FormatMarkdown,
		Content: content,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(cands) != 2 {
		t.Fatalf("candidates = %d, want 2: %#v", len(cands), cands)
	}
	for _, c := range cands {
		if c.Statement == "" {
			t.Fatal("empty statement")
		}
		if c.Start < 0 || c.End <= c.Start || c.End > len(content) {
			t.Fatalf("bad span [%d,%d) for %q", c.Start, c.End, c.Statement)
		}
		if content[c.Start:c.End] != c.Statement {
			t.Fatalf("span %q != statement %q", content[c.Start:c.End], c.Statement)
		}
	}
}

func TestEmittingCandidatesLeavesRTMUnchanged(t *testing.T) {
	dir := t.TempDir()
	rtm := filepath.Join(dir, "rtm_database.csv")
	before := "req_id,status\nREQ-EXISTING-001,PARTIAL\n"
	if err := os.WriteFile(rtm, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := prdextract.Extract(prdextract.Input{
		Format:  prdextract.FormatText,
		Content: "The system shall never write the RTM from extraction alone.",
	})
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(rtm)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != before {
		t.Fatalf("RTM changed:\nbefore=%q\nafter=%q", before, after)
	}
}



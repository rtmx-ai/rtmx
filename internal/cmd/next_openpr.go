package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// OpenPRLookup lists open pull requests. Tests inject fakes; production uses gh.
type OpenPRLookup interface {
	ListOpen() ([]PullRequest, error)
}

// openPRSource is nil in production (gh). Tests set a fake before runNext.
var openPRSource OpenPRLookup

type openPRSkip struct {
	ReqID  string `json:"req_id"`
	Reason string `json:"reason"`
}

type ghOpenPRLookup struct{}

func (ghOpenPRLookup) ListOpen() ([]PullRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	c := exec.CommandContext(ctx, "gh", "pr", "list", "--state", "open", "--json", "number,title,body", "--limit", "200")
	out, err := c.Output()
	if err != nil {
		return nil, fmt.Errorf("gh pr list: %w", err)
	}
	var rows []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}
	if err := json.Unmarshal(out, &rows); err != nil {
		return nil, fmt.Errorf("gh pr list json: %w", err)
	}
	prs := make([]PullRequest, 0, len(rows))
	for _, row := range rows {
		prs = append(prs, PullRequest{Number: row.Number, Title: row.Title, Body: row.Body})
	}
	return prs, nil
}

func lookupOpenPRNames(cmd *cobra.Command, ids []string) map[string]bool {
	src := openPRSource
	if src == nil {
		src = ghOpenPRLookup{}
	}
	prs, err := src.ListOpen()
	if err != nil {
		cmd.PrintErrf("warning: open pull request lookup unavailable (%v); not skipping\n", err)
		return map[string]bool{}
	}
	return namedByOpenPRs(ids, prs)
}

// namedByOpenPRs marks an ID when it appears as its own token in a PR title or body.
// A longer ID that merely starts with this one (REQ-ORCH-019 vs REQ-ORCH-019a) does not count.
func namedByOpenPRs(ids []string, prs []PullRequest) map[string]bool {
	named := map[string]bool{}
	for _, id := range ids {
		for _, pr := range prs {
			if textNamesReq(pr.Title, id) || textNamesReq(pr.Body, id) {
				named[id] = true
				break
			}
		}
	}
	return named
}

func textNamesReq(text, id string) bool {
	if id == "" || text == "" {
		return false
	}
	rest := text
	for {
		i := strings.Index(rest, id)
		if i < 0 {
			return false
		}
		end := i + len(id)
		beforeOK := i == 0 || !isReqIDByte(rest[i-1])
		afterOK := end == len(rest) || !isReqIDByte(rest[end])
		if beforeOK && afterOK {
			return true
		}
		rest = rest[i+1:]
	}
}

func isReqIDByte(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func openPRSkips(ids []string, named map[string]bool) (keep []string, skipped []openPRSkip) {
	skipped = []openPRSkip{}
	for _, id := range ids {
		if named[id] {
			skipped = append(skipped, openPRSkip{ReqID: id, Reason: "open_pr"})
			continue
		}
		keep = append(keep, id)
	}
	return keep, skipped
}

func printOpenPRSkips(cmd *cobra.Command, skipped []openPRSkip) {
	for _, s := range skipped {
		cmd.Printf("Skipped %s because of an open pull request\n", s.ReqID)
	}
}

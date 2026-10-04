// Package deliverycheck validates one-PR-per-REQ and one-commit-per-AC heuristics.
package deliverycheck

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var (
	reqIDPattern = regexp.MustCompile(`\bREQ-[A-Z0-9]+(?:-[A-Z0-9]+)*\b`)
	acPattern    = regexp.MustCompile(`(?i)\bAC-?\d+\b|\bac_id\s*=\s*\S+`)
)

// CommitInfo is one commit since the merge-base.
type CommitInfo struct {
	SHA     string `json:"sha"`
	Subject string `json:"subject"`
	OK      bool   `json:"ok"`
	Reason  string `json:"reason,omitempty"`
}

// Result is the delivery-check report.
type Result struct {
	OK       bool         `json:"ok"`
	ReqID    string       `json:"req_id,omitempty"`
	Warnings []string     `json:"warnings"`
	Errors   []string     `json:"errors"`
	Commits  []CommitInfo `json:"commits"`
}

// Input configures a check run.
type Input struct {
	// Dir is the git working tree (defaults to cwd of git commands).
	Dir string
	// Base is the merge-base ref (default "main").
	Base string
	// ReqID forces the requirement ID; empty means discover from Title+Body or commits.
	ReqID string
	// Title and Body are PR metadata; when both empty, commit subjects are used for REQ discovery.
	Title string
	Body  string
}

// Check runs warn-first delivery heuristics. Errors are populated for violations;
// callers decide whether to fail (--strict) or warn.
func Check(in Input) (*Result, error) {
	base := in.Base
	if base == "" {
		base = "main"
	}
	commits, err := listCommitsSinceBase(in.Dir, base)
	if err != nil {
		return nil, err
	}

	res := &Result{Warnings: []string{}, Errors: []string{}, Commits: []CommitInfo{}}

	reqID := strings.TrimSpace(in.ReqID)
	if reqID == "" {
		meta := in.Title + "\n" + in.Body
		if strings.TrimSpace(meta) == "" {
			var subjects []string
			for _, c := range commits {
				subjects = append(subjects, c.Subject)
			}
			meta = strings.Join(subjects, "\n")
		}
		ids := uniqueReqIDs(meta)
		switch len(ids) {
		case 0:
			res.Errors = append(res.Errors, "no REQ-… ID found in PR metadata or commit subjects")
		case 1:
			reqID = ids[0]
		default:
			res.Errors = append(res.Errors, fmt.Sprintf("expected exactly one REQ-… ID, found %d: %s", len(ids), strings.Join(ids, ", ")))
		}
	}
	res.ReqID = reqID

	if reqID != "" {
		meta := in.Title + "\n" + in.Body
		if strings.TrimSpace(meta) != "" {
			ids := uniqueReqIDs(meta)
			if len(ids) > 1 {
				res.Errors = append(res.Errors, fmt.Sprintf("PR metadata names multiple REQ IDs: %s", strings.Join(ids, ", ")))
			} else if len(ids) == 1 && ids[0] != reqID {
				res.Errors = append(res.Errors, fmt.Sprintf("PR metadata REQ %s does not match --req %s", ids[0], reqID))
			}
		}
	}

	if len(commits) == 0 {
		res.Warnings = append(res.Warnings, "no commits since merge-base with "+base)
	}

	for _, c := range commits {
		info := CommitInfo{SHA: c.SHA, Subject: c.Subject}
		if reqID == "" {
			info.OK = false
			info.Reason = "no requirement ID to match"
			res.Errors = append(res.Errors, fmt.Sprintf("commit %s: %s", shortSHA(c.SHA), info.Reason))
		} else if !strings.Contains(c.Subject, reqID) && !strings.Contains(c.Body, reqID) {
			info.OK = false
			info.Reason = "missing requirement ID " + reqID
			res.Errors = append(res.Errors, fmt.Sprintf("commit %s: %s", shortSHA(c.SHA), info.Reason))
		} else if !acPattern.MatchString(c.Subject) && !acPattern.MatchString(c.Body) {
			info.OK = false
			info.Reason = "missing AC marker (AC1, AC-1, or ac_id=…)"
			res.Errors = append(res.Errors, fmt.Sprintf("commit %s: %s", shortSHA(c.SHA), info.Reason))
		} else {
			info.OK = true
		}
		res.Commits = append(res.Commits, info)
	}

	res.OK = len(res.Errors) == 0
	if !res.OK {
		res.Warnings = append(res.Warnings, res.Errors...)
	}
	return res, nil
}

type rawCommit struct {
	SHA     string
	Subject string
	Body    string
}

func listCommitsSinceBase(dir, base string) ([]rawCommit, error) {
	// Resolve merge-base; if base missing, try origin/main then empty (all commits).
	mb, err := gitOutput(dir, "merge-base", "HEAD", base)
	if err != nil {
		if alt, altErr := gitOutput(dir, "merge-base", "HEAD", "origin/"+base); altErr == nil {
			mb = alt
		} else {
			// First commit only repo: compare against empty tree via --root style.
			out, logErr := gitOutput(dir, "log", "--format=%H%x00%s%x00%b%x1e", "HEAD")
			if logErr != nil {
				return nil, fmt.Errorf("git log: %w", err)
			}
			return parseCommitLog(out), nil
		}
	}
	rangeSpec := strings.TrimSpace(mb) + "..HEAD"
	out, err := gitOutput(dir, "log", "--format=%H%x00%s%x00%b%x1e", rangeSpec)
	if err != nil {
		return nil, fmt.Errorf("git log %s: %w", rangeSpec, err)
	}
	return parseCommitLog(out), nil
}

func parseCommitLog(out string) []rawCommit {
	var commits []rawCommit
	for _, entry := range strings.Split(out, "\x1e") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, "\x00", 3)
		if len(parts) < 2 {
			continue
		}
		c := rawCommit{SHA: parts[0], Subject: parts[1]}
		if len(parts) == 3 {
			c.Body = parts[2]
		}
		commits = append(commits, c)
	}
	// git log is newest-first; reverse to chronological for reporting.
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}
	return commits
}

func gitOutput(dir string, args ...string) (string, error) {
	c := exec.Command("git", args...)
	if dir != "" {
		c.Dir = dir
	}
	out, err := c.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func uniqueReqIDs(text string) []string {
	found := reqIDPattern.FindAllString(text, -1)
	seen := map[string]bool{}
	var out []string
	for _, id := range found {
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

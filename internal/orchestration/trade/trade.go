// Package trade stores trade-analysis checkpoint artifacts under .rtmx/trades/.
package trade

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var slugClean = regexp.MustCompile(`[^a-z0-9]+`)

// Status values.
const (
	StatusOpen     = "open"
	StatusResolved = "resolved"
)

// Trade is a parsed trade analysis file.
type Trade struct {
	ID       string `json:"id"`
	ReqID    string `json:"req_id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Choice   string `json:"choice,omitempty"`
	Path     string `json:"path"`
	Body     string `json:"-"`
}

// Dir returns .rtmx/trades under cwd.
func Dir(cwd string) string {
	return filepath.Join(cwd, ".rtmx", "trades")
}

// Open creates a new open trade for reqID.
func Open(cwd, reqID, title string) (*Trade, error) {
	if reqID == "" {
		return nil, fmt.Errorf("req_id is required")
	}
	if title == "" {
		title = "Trade analysis"
	}
	slug := slugify(title)
	id := fmt.Sprintf("TRADE-%s-%s", reqID, slug)
	dir := Dir(cwd)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, id+".md")
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("trade already exists: %s", id)
	}
	body := fmt.Sprintf(`---
id: %s
req_id: %s
title: %q
status: open
created: %s
---

# %s

## Options

- Option A:
- Option B:

## Criteria

-

## Recommendation

-

## Status

open
`, id, reqID, title, time.Now().UTC().Format(time.RFC3339), title)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return nil, err
	}
	return &Trade{ID: id, ReqID: reqID, Title: title, Status: StatusOpen, Path: path, Body: body}, nil
}

// List returns trades, optionally filtered by reqID.
func List(cwd, reqID string) ([]Trade, error) {
	dir := Dir(cwd)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Trade{}, nil
		}
		return nil, err
	}
	var out []Trade
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		t, err := ParseFile(path)
		if err != nil {
			continue
		}
		if reqID != "" && t.ReqID != reqID {
			continue
		}
		out = append(out, *t)
	}
	return out, nil
}

// Resolve marks a trade resolved with a choice.
func Resolve(cwd, tradeID, choice string) (*Trade, error) {
	if tradeID == "" {
		return nil, fmt.Errorf("trade id is required")
	}
	if choice == "" {
		return nil, fmt.Errorf("choice is required")
	}
	path := filepath.Join(Dir(cwd), tradeID+".md")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("trade %s: %w", tradeID, err)
	}
	body := string(raw)
	body = replaceFrontMatter(body, "status", StatusResolved)
	body = upsertFrontMatter(body, "choice", choice)
	body = replaceStatusSection(body, StatusResolved)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return nil, err
	}
	t, err := ParseFile(path)
	if err != nil {
		return nil, err
	}
	t.Choice = choice
	return t, nil
}

// HasOpen returns true if any open trade exists for reqID.
func HasOpen(cwd, reqID string) (bool, error) {
	trades, err := List(cwd, reqID)
	if err != nil {
		return false, err
	}
	for _, t := range trades {
		if t.Status == StatusOpen {
			return true, nil
		}
	}
	return false, nil
}

// NeedsTrade reports whether requirement text/body suggests a trade is required.
func NeedsTrade(reqText, mdBody string) bool {
	blob := strings.ToLower(reqText + "\n" + mdBody)
	if strings.Contains(blob, "needs_trade: true") || strings.Contains(blob, "needs_trade:true") {
		return true
	}
	markers := []string{"tbd", "todo(decision)", "either/or", "options:"}
	for _, m := range markers {
		if strings.Contains(blob, m) {
			return true
		}
	}
	return false
}

// ParseFile reads a trade markdown file.
func ParseFile(path string) (*Trade, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	body := string(raw)
	t := &Trade{Path: path, Body: body, Status: StatusOpen}
	t.ID = fmValue(body, "id")
	if t.ID == "" {
		t.ID = strings.TrimSuffix(filepath.Base(path), ".md")
	}
	t.ReqID = fmValue(body, "req_id")
	t.Title = strings.Trim(fmValue(body, "title"), `"'`)
	if st := fmValue(body, "status"); st != "" {
		t.Status = st
	}
	t.Choice = fmValue(body, "choice")
	if t.Status == "" {
		t.Status = StatusOpen
	}
	return t, nil
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	slug := slugClean.ReplaceAllString(b.String(), "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "trade"
	}
	if len(slug) > 40 {
		slug = slug[:40]
		slug = strings.Trim(slug, "-")
	}
	return slug
}

func fmValue(body, key string) string {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `:\s*(.+)$`)
	m := re.FindStringSubmatch(body)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func replaceFrontMatter(body, key, value string) string {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `:\s*.+$`)
	if re.MatchString(body) {
		return re.ReplaceAllString(body, key+": "+value)
	}
	return upsertFrontMatter(body, key, value)
}

func upsertFrontMatter(body, key, value string) string {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `:\s*.+$`)
	if re.MatchString(body) {
		return re.ReplaceAllString(body, key+": "+value)
	}
	// Insert after opening ---
	if strings.HasPrefix(body, "---\n") {
		return "---\n" + key + ": " + value + "\n" + strings.TrimPrefix(body, "---\n")
	}
	return body + "\n" + key + ": " + value + "\n"
}

func replaceStatusSection(body, status string) string {
	re := regexp.MustCompile(`(?ms)(## Status\s*\n)[^\n]*`)
	if re.MatchString(body) {
		return re.ReplaceAllString(body, "${1}"+status)
	}
	return body
}

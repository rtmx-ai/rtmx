package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestMCPNextDeliveryPlan(t *testing.T) {
	rtmx.Req(t, "REQ-MCP-012")

	bin := t.TempDir()
	marker := filepath.Join(bin, "gh-called")
	script := filepath.Join(bin, "gh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\ntouch \"$GH_MARKER\"\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GH_MARKER", marker)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	t.Run("keeps_webs_and_plans_atomic", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, ".rtmx", "database.csv")
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
			t.Fatal(err)
		}
		db := database.NewDatabase()
		req := &database.Requirement{
			ReqID:           "REQ-EX-070",
			Category:        "CLI",
			RequirementText: "Atomic item",
			Status:          database.StatusMissing,
			Priority:        database.PriorityHigh,
			RequirementFile: ".rtmx/requirements/CLI/REQ-EX-070.md",
		}
		if err := db.Add(req); err != nil {
			t.Fatal(err)
		}
		if err := db.Save(dbPath); err != nil {
			t.Fatal(err)
		}
		md := filepath.Join(dir, ".rtmx", "requirements", "CLI", "REQ-EX-070.md")
		if err := os.MkdirAll(filepath.Dir(md), 0o755); err != nil {
			t.Fatal(err)
		}
		body := "# REQ-EX-070\n\n## Acceptance Criteria\n\n- [ ] one\n- [ ] two\n"
		if err := os.WriteFile(md, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		srv := NewServer(dbPath, config.DefaultConfig())
		loaded, err := database.Load(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		before := len(loaded.All())
		raw, err := json.Marshal(srv.toolNext(loaded, toolFilter{}))
		if err != nil {
			t.Fatal(err)
		}
		var parsed struct {
			Webs []struct {
				TopItem string `json:"top_item"`
			} `json:"webs"`
			Delivery struct {
				ReqID        string   `json:"req_id"`
				Decomposed   bool     `json:"decomposed"`
				Children     []string `json:"children"`
				PRPolicy     string   `json:"pr_policy"`
				CommitPolicy string   `json:"commit_policy"`
				Skipped      []struct {
					ReqID  string `json:"req_id"`
					Reason string `json:"reason"`
				} `json:"skipped"`
			} `json:"delivery"`
		}
		if err := json.Unmarshal(raw, &parsed); err != nil {
			t.Fatal(err)
		}
		if len(parsed.Webs) == 0 || parsed.Webs[0].TopItem != "REQ-EX-070" {
			t.Fatalf("webs/top_item changed: %s", raw)
		}
		if parsed.Delivery.ReqID != "REQ-EX-070" || parsed.Delivery.Decomposed || len(parsed.Delivery.Children) != 0 {
			t.Fatalf("atomic plan: %s", raw)
		}
		if parsed.Delivery.PRPolicy != "one_pr_per_requirement" || parsed.Delivery.CommitPolicy != "one_commit_per_ac" {
			t.Fatalf("policies: %s", raw)
		}
		if parsed.Delivery.Skipped == nil {
			t.Fatalf("skipped must be an array: %s", raw)
		}
		after, err := database.Load(dbPath)
		if err != nil || len(after.All()) != before {
			t.Fatalf("next mutated the database: %v", err)
		}
	})

	t.Run("coarse_lists_children", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, ".rtmx", "database.csv")
		_ = os.MkdirAll(filepath.Dir(dbPath), 0o755)
		db := database.NewDatabase()
		req := &database.Requirement{
			ReqID:           "REQ-EX-071",
			Category:        "CLI",
			RequirementText: "Coarse item",
			Status:          database.StatusMissing,
			Priority:        database.PriorityHigh,
			RequirementFile: ".rtmx/requirements/CLI/REQ-EX-071.md",
		}
		if err := db.Add(req); err != nil {
			t.Fatal(err)
		}
		if err := db.Save(dbPath); err != nil {
			t.Fatal(err)
		}
		md := filepath.Join(dir, ".rtmx", "requirements", "CLI", "REQ-EX-071.md")
		_ = os.MkdirAll(filepath.Dir(md), 0o755)
		body := "# REQ-EX-071\n\n## Acceptance Criteria\n\n- [ ] a\n- [ ] b\n- [ ] c\n- [ ] d\n- [ ] e\n- [ ] f\n"
		if err := os.WriteFile(md, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		srv := NewServer(dbPath, config.DefaultConfig())
		loaded, err := database.Load(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(srv.toolNext(loaded, toolFilter{}))
		var parsed struct {
			Delivery struct {
				ReqID      string   `json:"req_id"`
				Decomposed bool     `json:"decomposed"`
				Children   []string `json:"children"`
			} `json:"delivery"`
		}
		if err := json.Unmarshal(raw, &parsed); err != nil {
			t.Fatal(err)
		}
		if parsed.Delivery.ReqID != "REQ-EX-071" || !parsed.Delivery.Decomposed || len(parsed.Delivery.Children) != 6 {
			t.Fatalf("coarse plan: %s", raw)
		}
		if parsed.Delivery.Children[0] != "REQ-EX-071a" {
			t.Fatalf("child ids: %v", parsed.Delivery.Children)
		}
	})

	t.Run("idle_does_not_claim", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, ".rtmx", "database.csv")
		_ = os.MkdirAll(filepath.Dir(dbPath), 0o755)
		db := database.NewDatabase()
		req := &database.Requirement{
			ReqID:    "REQ-EX-072",
			Category: "CLI",
			Status:   database.StatusComplete,
			Priority: database.PriorityHigh,
		}
		if err := db.Add(req); err != nil {
			t.Fatal(err)
		}
		if err := db.Save(dbPath); err != nil {
			t.Fatal(err)
		}
		srv := NewServer(dbPath, config.DefaultConfig())
		loaded, err := database.Load(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(srv.toolNext(loaded, toolFilter{}))
		var parsed struct {
			Delivery struct {
				Idle bool `json:"idle"`
			} `json:"delivery"`
		}
		if err := json.Unmarshal(raw, &parsed); err != nil {
			t.Fatalf("json: %v %s", err, raw)
		}
		if !parsed.Delivery.Idle {
			t.Fatalf("expected idle: %s", raw)
		}
		claims, _ := filepath.Glob(filepath.Join(dir, ".rtmx", "claims", "*.json"))
		if len(claims) != 0 {
			t.Fatalf("idle claimed: %v", claims)
		}
	})

	if _, err := os.Stat(marker); err == nil {
		t.Fatal("next called gh")
	}
}

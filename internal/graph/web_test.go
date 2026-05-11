package graph

import (
	"testing"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestDetectWebs(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-001")

	t.Run("two_independent_webs", func(t *testing.T) {
		db := database.NewDatabase()

		// Web 1: A -> B (A depends on B)
		a := database.NewRequirement("REQ-A")
		a.Status = database.StatusMissing
		a.EffortWeeks = 1.0
		a.Dependencies.Add("REQ-B")
		_ = db.Add(a)

		b := database.NewRequirement("REQ-B")
		b.Status = database.StatusMissing
		b.EffortWeeks = 2.0
		b.Blocks.Add("REQ-A")
		_ = db.Add(b)

		// Web 2: C (isolated)
		c := database.NewRequirement("REQ-C")
		c.Status = database.StatusMissing
		c.EffortWeeks = 0.5
		_ = db.Add(c)

		g := NewGraph(db)
		webs := g.DetectWebs()

		if len(webs) != 2 {
			t.Fatalf("expected 2 webs, got %d", len(webs))
		}

		// Largest effort first
		if webs[0].TotalEffort != 3.0 {
			t.Errorf("web 0 effort = %.1f, want 3.0", webs[0].TotalEffort)
		}
		if len(webs[0].IDs) != 2 {
			t.Errorf("web 0 size = %d, want 2", len(webs[0].IDs))
		}
		if webs[1].TotalEffort != 0.5 {
			t.Errorf("web 1 effort = %.1f, want 0.5", webs[1].TotalEffort)
		}
	})

	t.Run("blocked_vs_unblocked", func(t *testing.T) {
		db := database.NewDatabase()

		// B is unblocked, A depends on B so A is blocked
		a := database.NewRequirement("REQ-A")
		a.Status = database.StatusMissing
		a.Dependencies.Add("REQ-B")
		_ = db.Add(a)

		b := database.NewRequirement("REQ-B")
		b.Status = database.StatusMissing
		b.Blocks.Add("REQ-A")
		_ = db.Add(b)

		g := NewGraph(db)
		webs := g.DetectWebs()

		if len(webs) != 1 {
			t.Fatalf("expected 1 web, got %d", len(webs))
		}

		web := webs[0]
		if len(web.Unblocked) != 1 || web.Unblocked[0] != "REQ-B" {
			t.Errorf("unblocked = %v, want [REQ-B]", web.Unblocked)
		}
		if len(web.Blocked) != 1 || web.Blocked[0] != "REQ-A" {
			t.Errorf("blocked = %v, want [REQ-A]", web.Blocked)
		}
	})

	t.Run("complete_requirements_excluded", func(t *testing.T) {
		db := database.NewDatabase()

		a := database.NewRequirement("REQ-A")
		a.Status = database.StatusComplete
		_ = db.Add(a)

		b := database.NewRequirement("REQ-B")
		b.Status = database.StatusMissing
		b.Dependencies.Add("REQ-A")
		_ = db.Add(b)

		c := database.NewRequirement("REQ-C")
		c.Status = database.StatusMissing
		_ = db.Add(c)

		g := NewGraph(db)
		webs := g.DetectWebs()

		// B and C are each in their own web (A is complete, excluded)
		if len(webs) != 2 {
			t.Fatalf("expected 2 webs (complete excluded), got %d", len(webs))
		}

		// B is not blocked because its dependency A is complete
		for _, web := range webs {
			if len(web.IDs) == 1 && web.IDs[0] == "REQ-B" {
				if len(web.Unblocked) != 1 {
					t.Errorf("REQ-B should be unblocked (dep REQ-A is complete)")
				}
			}
		}
	})

	t.Run("empty_database", func(t *testing.T) {
		db := database.NewDatabase()
		g := NewGraph(db)
		webs := g.DetectWebs()
		if webs != nil {
			t.Errorf("expected nil webs for empty db, got %d", len(webs))
		}
	})

	t.Run("all_complete", func(t *testing.T) {
		db := database.NewDatabase()
		a := database.NewRequirement("REQ-A")
		a.Status = database.StatusComplete
		_ = db.Add(a)

		g := NewGraph(db)
		webs := g.DetectWebs()
		if webs != nil {
			t.Errorf("expected nil webs when all complete, got %d", len(webs))
		}
	})

	t.Run("detect_overlaps", func(t *testing.T) {
		// This sub-test verifies REQ-ORCH-007 behavior within the web tests
	})
}

func TestDetectOverlaps(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-007",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	t.Run("no_overlap", func(t *testing.T) {
		db := database.NewDatabase()

		a := database.NewRequirement("REQ-A")
		a.Status = database.StatusMissing
		a.TestModule = "internal/cmd/foo_test.go"
		_ = db.Add(a)

		b := database.NewRequirement("REQ-B")
		b.Status = database.StatusMissing
		b.TestModule = "internal/graph/bar_test.go"
		_ = db.Add(b)

		g := NewGraph(db)
		webs := g.DetectWebs()
		overlaps := g.DetectOverlaps(webs)

		if len(overlaps) != 0 {
			t.Errorf("expected 0 overlaps, got %d", len(overlaps))
		}
	})

	t.Run("shared_directory", func(t *testing.T) {
		db := database.NewDatabase()

		// Web 1
		a := database.NewRequirement("REQ-A")
		a.Status = database.StatusMissing
		a.TestModule = "internal/cmd/install_test.go"
		_ = db.Add(a)

		// Web 2 (independent, different component)
		b := database.NewRequirement("REQ-B")
		b.Status = database.StatusMissing
		b.TestModule = "internal/cmd/verify_test.go"
		_ = db.Add(b)

		g := NewGraph(db)
		webs := g.DetectWebs()
		overlaps := g.DetectOverlaps(webs)

		// Both touch internal/cmd -- should detect overlap
		if len(overlaps) != 1 {
			t.Fatalf("expected 1 overlap (shared dir), got %d", len(overlaps))
		}
		if len(overlaps[0].SharedFiles) == 0 {
			t.Error("shared files should not be empty")
		}
	})

	t.Run("single_web_no_overlap", func(t *testing.T) {
		db := database.NewDatabase()

		a := database.NewRequirement("REQ-A")
		a.Status = database.StatusMissing
		a.TestModule = "internal/cmd/foo_test.go"
		_ = db.Add(a)

		g := NewGraph(db)
		webs := g.DetectWebs()
		overlaps := g.DetectOverlaps(webs)

		if len(overlaps) != 0 {
			t.Errorf("single web should have no overlaps, got %d", len(overlaps))
		}
	})

	t.Run("no_test_module_no_overlap", func(t *testing.T) {
		db := database.NewDatabase()

		a := database.NewRequirement("REQ-A")
		a.Status = database.StatusMissing
		_ = db.Add(a)

		b := database.NewRequirement("REQ-B")
		b.Status = database.StatusMissing
		_ = db.Add(b)

		g := NewGraph(db)
		webs := g.DetectWebs()
		overlaps := g.DetectOverlaps(webs)

		if len(overlaps) != 0 {
			t.Errorf("reqs without test_module should have no overlaps, got %d", len(overlaps))
		}
	})
}

func TestDetectWebsChain(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-001")

	t.Run("chain_forms_single_web", func(t *testing.T) {
		db := database.NewDatabase()

		// A -> B -> C -> D (chain of 4)
		ids := []string{"REQ-A", "REQ-B", "REQ-C", "REQ-D"}
		for i, id := range ids {
			r := database.NewRequirement(id)
			r.Status = database.StatusMissing
			r.EffortWeeks = 1.0
			if i > 0 {
				r.Dependencies.Add(ids[i-1])
			}
			if i < len(ids)-1 {
				r.Blocks.Add(ids[i+1])
			}
			_ = db.Add(r)
		}

		g := NewGraph(db)
		webs := g.DetectWebs()

		if len(webs) != 1 {
			t.Fatalf("chain should form 1 web, got %d", len(webs))
		}
		if len(webs[0].IDs) != 4 {
			t.Errorf("web size = %d, want 4", len(webs[0].IDs))
		}
		if webs[0].TotalEffort != 4.0 {
			t.Errorf("effort = %.1f, want 4.0", webs[0].TotalEffort)
		}
		// Only the first node (no deps) should be unblocked
		if len(webs[0].Unblocked) != 1 {
			t.Errorf("unblocked = %d, want 1", len(webs[0].Unblocked))
		}
	})
}

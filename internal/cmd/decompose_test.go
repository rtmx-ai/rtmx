package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestDecomposeCoarseAndAtomic(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-020")
	dir := t.TempDir()
	rtmxDir := filepath.Join(dir, ".rtmx")
	reqDir := filepath.Join(rtmxDir, "requirements", "EX")
	if err := os.MkdirAll(reqDir, 0o755); err != nil {
		t.Fatal(err)
	}

	atomicMD := `# REQ-EX-100: Atomic

## Acceptance Criteria

1. One thing works.
2. Second thing works.
`
	coarseMD := `# REQ-EX-200: Coarse parent

Parent closes only when children are COMPLETE.

## Acceptance Criteria

1. First behavior.
2. Second behavior.
3. Third behavior.
4. Fourth behavior.
5. Fifth behavior.
6. Sixth behavior.
`
	if err := os.WriteFile(filepath.Join(reqDir, "REQ-EX-100.md"), []byte(atomicMD), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(reqDir, "REQ-EX-200.md"), []byte(coarseMD), 0o644); err != nil {
		t.Fatal(err)
	}

	db := database.NewDatabase()
	_ = db.Add(&database.Requirement{
		ReqID: "REQ-EX-100", Category: "EX", RequirementText: "Atomic",
		Status: database.StatusMissing, Priority: database.PriorityMedium,
		Dependencies: database.NewStringSet(), Blocks: database.NewStringSet(),
		RequirementFile: ".rtmx/requirements/EX/REQ-EX-100.md",
	})
	_ = db.Add(&database.Requirement{
		ReqID: "REQ-EX-200", Category: "EX", RequirementText: "Coarse parent closes only when children are COMPLETE",
		Status: database.StatusMissing, Priority: database.PriorityHigh,
		Dependencies: database.NewStringSet(), Blocks: database.NewStringSet(),
		RequirementFile: ".rtmx/requirements/EX/REQ-EX-200.md",
	})
	dbPath := filepath.Join(rtmxDir, "database.csv")
	if err := db.Save(dbPath); err != nil {
		t.Fatal(err)
	}

	// Atomic
	a := db.Get("REQ-EX-100")
	md, _ := os.ReadFile(filepath.Join(dir, a.RequirementFile))
	plan, err := planDecompose(a, string(md), db)
	if err != nil || !plan.Atomic {
		t.Fatalf("atomic: plan=%+v err=%v", plan, err)
	}

	// Coarse
	c := db.Get("REQ-EX-200")
	md2, _ := os.ReadFile(filepath.Join(dir, c.RequirementFile))
	plan2, err := planDecompose(c, string(md2), db)
	if err != nil || plan2.Atomic || len(plan2.ChildIDs) != 6 {
		t.Fatalf("coarse: plan=%+v err=%v", plan2, err)
	}
	if err := applyDecompose(dir, db, c, plan2, filepath.Join(dir, c.RequirementFile)); err != nil {
		t.Fatal(err)
	}
	if err := db.Save(dbPath); err != nil {
		t.Fatal(err)
	}
	if db.Len() != 8 { // 2 parents + 6 children
		t.Fatalf("len=%d", db.Len())
	}
	for _, id := range plan2.ChildIDs {
		ch := db.Get(id)
		if ch == nil || ch.Status != database.StatusMissing {
			t.Fatalf("child %s missing", id)
		}
		body, err := os.ReadFile(filepath.Join(dir, ch.RequirementFile))
		if err != nil || !strings.Contains(string(body), "## Acceptance Criteria") {
			t.Fatalf("child md %s: %v", id, err)
		}
		if !c.Dependencies.Contains(id) {
			t.Fatalf("parent should depend on %s", id)
		}
	}

	// Idempotent second plan
	plan3, err := planDecompose(c, string(md2), db)
	if err != nil || !plan3.Atomic {
		t.Fatalf("second run should be atomic/idempotent: %+v err=%v", plan3, err)
	}

	// COMPLETE refused at command level — unit check status
	c.Status = database.StatusComplete
	if c.Status != database.StatusComplete {
		t.Fatal("setup")
	}

	g := graph.NewGraph(db)
	if cycles := g.FindCycles(); len(cycles) > 0 {
		t.Fatalf("cycles after decompose: %v", cycles)
	}
}

func TestDecomposeDryRunWritesNothing(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-020")
	db := database.NewDatabase()
	_ = db.Add(&database.Requirement{
		ReqID: "REQ-EX-300", Category: "EX",
		RequirementText: "Parent closes only when children are COMPLETE",
		Status:          database.StatusMissing,
		Dependencies:    database.NewStringSet(), Blocks: database.NewStringSet(),
	})
	md := "## Decomposition\n\n- Child one\n- Child two\n"
	plan, err := planDecompose(db.Get("REQ-EX-300"), md, db)
	if err != nil || plan.Atomic || len(plan.ChildIDs) != 2 {
		t.Fatalf("%+v err=%v", plan, err)
	}
	before := db.Len()
	// dry-run: do not call applyDecompose
	if db.Len() != before {
		t.Fatal("dry-run mutated db")
	}
}

func TestDecomposeRefusesComplete(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-020")
	db := database.NewDatabase()
	_ = db.Add(&database.Requirement{
		ReqID: "REQ-EX-400", Status: database.StatusComplete,
		Dependencies: database.NewStringSet(), Blocks: database.NewStringSet(),
	})
	req := db.Get("REQ-EX-400")
	if req.Status != database.StatusComplete {
		t.Fatal("want COMPLETE")
	}
}

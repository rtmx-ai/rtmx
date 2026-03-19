package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/internal/graph"
	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
	"github.com/spf13/cobra"
)

func TestBacklogRealCommand(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createBacklogTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"backlog"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("backlog command failed: %v", err)
	}

	output := buf.String()
	expectedPhrases := []string{
		"Prioritized Backlog",
		"Total Requirements:",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected output to contain %q, got:\n%s", phrase, output)
		}
	}
}

func TestBacklogPhaseFilter(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createBacklogTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"backlog", "--phase", "1"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("backlog --phase 1 failed: %v", err)
	}

	output := buf.String()
	// Phase 1 is complete, so backlog should be empty or show no items
	if !strings.Contains(output, "Prioritized Backlog") {
		t.Errorf("Expected backlog header, got:\n%s", output)
	}
}

func TestBacklogViewModes(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	views := []string{"all", "critical", "quick-wins", "blockers", "list"}

	for _, view := range views {
		t.Run(view, func(t *testing.T) {
			rootCmd := createBacklogTestCmd()
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetArgs([]string{"backlog", "--view", view})

			err := rootCmd.Execute()
			if err != nil {
				t.Fatalf("backlog --view %s failed: %v", view, err)
			}

			output := buf.String()
			if !strings.Contains(output, "Prioritized Backlog") {
				t.Errorf("Expected backlog output for view %s, got:\n%s", view, output)
			}
		})
	}
}

// TestBacklogTableFormat verifies that backlog uses ASCII table format
// REQ-GO-048: Go CLI shall use ASCII tables matching Python tabulate output
func TestBacklogTableFormat(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createBacklogTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"backlog"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("backlog command failed: %v", err)
	}

	output := buf.String()

	// Verify ASCII table format markers
	expectedTableElements := []string{
		"+---+",      // Column separator
		"|",          // Row borders
		"+===+",      // Header separator (with = instead of -)
		"Status",     // Column header
		"Requirement", // Column header
		"Description", // Column header
	}

	for _, element := range expectedTableElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected table format element %q, got:\n%s", element, output)
		}
	}

	// Verify sections exist
	expectedSections := []string{
		"Prioritized Backlog",
		"Total Requirements:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected section %q, got:\n%s", section, output)
		}
	}
}

// createBacklogTestDB creates a synthetic database for testing transitive blocking.
// Dependency graph:
//
//	A (incomplete) blocks B, C
//	B (incomplete) blocks D
//	C (complete) blocks D
//	D (incomplete) depends on B, C
//	E (incomplete) no dependencies
//
// Transitive blocking counts:
//
//	A blocks B, D transitively (2 transitive), direct dependents B, C (but C is complete, so 1 direct blocker counted)
//	B blocks D transitively (1 transitive), 1 direct
//	C is complete, blocks nothing incomplete
//	D blocks nothing
//	E blocks nothing
//
// Quick-wins candidates: must be high priority, low effort, AND not blocked by incomplete deps.
func createBacklogTestDB() *database.Database {
	db := database.NewDatabase()

	reqs := []*database.Requirement{
		{
			ReqID: "REQ-A-001", Category: "TEST", Status: database.StatusMissing,
			Priority: database.PriorityHigh, Phase: 1, EffortWeeks: 0.5,
			RequirementText: "Root requirement A",
			Dependencies:    database.NewStringSet(),
			Blocks:          database.NewStringSet("REQ-B-001", "REQ-C-001"),
		},
		{
			ReqID: "REQ-B-001", Category: "TEST", Status: database.StatusMissing,
			Priority: database.PriorityHigh, Phase: 1, EffortWeeks: 0.5,
			RequirementText: "Middle requirement B",
			Dependencies:    database.NewStringSet("REQ-A-001"),
			Blocks:          database.NewStringSet("REQ-D-001"),
		},
		{
			ReqID: "REQ-C-001", Category: "TEST", Status: database.StatusComplete,
			Priority: database.PriorityMedium, Phase: 1, EffortWeeks: 0.5,
			RequirementText: "Completed requirement C",
			Dependencies:    database.NewStringSet("REQ-A-001"),
			Blocks:          database.NewStringSet("REQ-D-001"),
		},
		{
			ReqID: "REQ-D-001", Category: "TEST", Status: database.StatusMissing,
			Priority: database.PriorityHigh, Phase: 2, EffortWeeks: 0.5,
			RequirementText: "Leaf requirement D",
			Dependencies:    database.NewStringSet("REQ-B-001", "REQ-C-001"),
			Blocks:          database.NewStringSet(),
		},
		{
			ReqID: "REQ-E-001", Category: "TEST", Status: database.StatusMissing,
			Priority: database.PriorityHigh, Phase: 1, EffortWeeks: 0.5,
			RequirementText: "Independent requirement E",
			Dependencies:    database.NewStringSet(),
			Blocks:          database.NewStringSet(),
		},
	}

	for _, r := range reqs {
		r.Extra = make(map[string]string)
		_ = db.Add(r)
	}

	return db
}

// TestBacklogTransitiveBlocking verifies that blocking counts use transitive closure
// and that the Blocks column shows "X (Y)" format where X=transitive, Y=direct.
func TestBacklogTransitiveBlocking(t *testing.T) {
	rtmx.Req(t, "REQ-PAR-003", rtmx.Scope("unit"), rtmx.Technique("nominal"))

	db := createBacklogTestDB()
	g := graph.NewGraph(db)

	// Test countBlockedTransitive: A should transitively block 2 incomplete reqs (B and D)
	transitiveA := countBlockedTransitive("REQ-A-001", g)
	if transitiveA != 2 {
		t.Errorf("A should transitively block 2 incomplete reqs, got %d", transitiveA)
	}

	// Direct blocking: A directly blocks B (incomplete) and C (complete) = 1 direct incomplete
	directA := countBlocked(&database.Requirement{ReqID: "REQ-A-001"}, db)
	if directA != 1 {
		t.Errorf("A should directly block 1 incomplete req (B), got %d", directA)
	}

	// B should transitively block 1 (D)
	transitiveB := countBlockedTransitive("REQ-B-001", g)
	if transitiveB != 1 {
		t.Errorf("B should transitively block 1 incomplete req, got %d", transitiveB)
	}

	// E blocks nothing
	transitiveE := countBlockedTransitive("REQ-E-001", g)
	if transitiveE != 0 {
		t.Errorf("E should transitively block 0 reqs, got %d", transitiveE)
	}
}

// TestBacklogQuickWinsExcludesBlocked verifies that quick-wins excludes
// requirements that are blocked by incomplete dependencies.
func TestBacklogQuickWinsExcludesBlocked(t *testing.T) {
	rtmx.Req(t, "REQ-PAR-003", rtmx.Scope("unit"), rtmx.Technique("nominal"))

	db := createBacklogTestDB()
	reqs := db.Incomplete()

	quickWins := filterQuickWins(reqs, db)

	// REQ-B-001 is HIGH priority, 0.5 effort, but blocked by A -> excluded
	// REQ-D-001 is HIGH priority, 0.5 effort, but blocked by B -> excluded
	// REQ-A-001 is HIGH priority, 0.5 effort, not blocked -> included
	// REQ-E-001 is HIGH priority, 0.5 effort, not blocked -> included

	for _, qw := range quickWins {
		if qw.ReqID == "REQ-B-001" {
			t.Error("quick-wins should NOT include REQ-B-001 (blocked by REQ-A-001)")
		}
		if qw.ReqID == "REQ-D-001" {
			t.Error("quick-wins should NOT include REQ-D-001 (blocked by REQ-B-001)")
		}
	}

	// Should include the unblocked ones
	foundA := false
	foundE := false
	for _, qw := range quickWins {
		if qw.ReqID == "REQ-A-001" {
			foundA = true
		}
		if qw.ReqID == "REQ-E-001" {
			foundE = true
		}
	}
	if !foundA {
		t.Error("quick-wins should include REQ-A-001 (high priority, low effort, not blocked)")
	}
	if !foundE {
		t.Error("quick-wins should include REQ-E-001 (high priority, low effort, not blocked)")
	}

	if len(quickWins) != 2 {
		t.Errorf("Expected 2 quick-wins, got %d", len(quickWins))
	}
}

// TestBacklogBlocksColumnFormat verifies the "X (Y)" format in the Blocks column.
func TestBacklogBlocksColumnFormat(t *testing.T) {
	rtmx.Req(t, "REQ-PAR-003", rtmx.Scope("unit"), rtmx.Technique("nominal"))

	db := createBacklogTestDB()
	g := graph.NewGraph(db)

	req := db.Get("REQ-A-001")
	transitive := countBlockedTransitive(req.ReqID, g)
	direct := countBlocked(req, db)
	blocksStr := formatBlocksColumn(transitive, direct)

	// A: transitive=2, direct=1 -> "2 (1)"
	if blocksStr != "2 (1)" {
		t.Errorf("Expected blocks column '2 (1)', got %q", blocksStr)
	}

	// E: transitive=0, direct=0 -> "0 (0)"
	reqE := db.Get("REQ-E-001")
	transitiveE := countBlockedTransitive(reqE.ReqID, g)
	directE := countBlocked(reqE, db)
	blocksStrE := formatBlocksColumn(transitiveE, directE)
	if blocksStrE != "0 (0)" {
		t.Errorf("Expected blocks column '0 (0)', got %q", blocksStrE)
	}
}

// TestBacklogJSON verifies that --json produces valid JSON with correct structure.
// REQ-PAR-001: JSON output flag for backlog command
func TestBacklogJSON(t *testing.T) {
	rtmx.Req(t, "REQ-PAR-001", rtmx.Scope("unit"), rtmx.Technique("nominal"))

	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createBacklogTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"backlog", "--json"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("backlog --json failed: %v", err)
	}

	output := buf.String()

	// Must be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput:\n%s", err, output)
	}

	// Verify required top-level fields
	requiredFields := []string{"total_missing", "estimated_effort_weeks", "remaining"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("JSON output missing required field %q", field)
		}
	}

	// Verify remaining is an array with correct structure
	remaining, ok := result["remaining"].([]interface{})
	if !ok {
		t.Fatalf("remaining is not an array")
	}
	if len(remaining) > 0 {
		item := remaining[0].(map[string]interface{})
		for _, field := range []string{"req_id", "description", "priority", "blocked"} {
			if _, ok := item[field]; !ok {
				t.Errorf("remaining entry missing required field %q", field)
			}
		}
	}

	// Verify no human-readable text
	outputStr := strings.TrimSpace(output)
	if strings.Contains(outputStr, "Prioritized Backlog") {
		t.Error("JSON output should not contain 'Prioritized Backlog' header")
	}
}

// createBacklogTestCmd creates a root command with real backlog command for testing
func createBacklogTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var view string
	var phase int
	var category string
	var limit int
	var jsonOutput bool

	backlogCmd := &cobra.Command{
		Use:   "backlog",
		Short: "Show prioritized backlog",
		RunE: func(cmd *cobra.Command, args []string) error {
			backlogView = view
			backlogPhase = phase
			backlogCategory = category
			backlogLimit = limit
			backlogJSON = jsonOutput
			return runBacklog(cmd, args)
		},
	}
	backlogCmd.Flags().StringVar(&view, "view", "all", "view mode")
	backlogCmd.Flags().IntVar(&phase, "phase", 0, "filter by phase")
	backlogCmd.Flags().StringVar(&category, "category", "", "filter by category")
	backlogCmd.Flags().IntVarP(&limit, "limit", "n", 0, "limit results")
	backlogCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	root.AddCommand(backlogCmd)

	return root
}

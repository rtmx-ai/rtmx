package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/internal/graph"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var cyclesJSON bool

var cyclesCmd = &cobra.Command{
	Use:   "cycles",
	Short: "Detect circular dependencies",
	Long: `Detect and display circular dependencies in the requirement graph.

Circular dependencies prevent requirements from ever being completed
because they form a deadlock where each requirement waits for another.

Uses Tarjan's Strongly Connected Components algorithm for detection.

Exit codes:
  0  No cycles found
  1  Cycles found`,
	RunE: runCycles,
}

func init() {
	cyclesCmd.Flags().BoolVar(&cyclesJSON, "json", false, "output as JSON")

	rootCmd.AddCommand(cyclesCmd)
}

func runCycles(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	dbPath := cfg.DatabasePath(cwd)
	db, err := database.Load(dbPath)
	if err != nil {
		return fmt.Errorf("failed to load database: %w", err)
	}

	g := graph.NewGraph(db)
	cycles := g.FindCycles()

	if cyclesJSON {
		return outputCyclesJSON(cmd, cycles, g)
	}
	return outputCyclesText(cmd, cycles, g, db)
}

type cycleResult struct {
	Found  bool       `json:"found"`
	Count  int        `json:"count"`
	Cycles [][]string `json:"cycles"`
}

func outputCyclesJSON(cmd *cobra.Command, cycles [][]string, g *graph.Graph) error {
	result := cycleResult{
		Found:  len(cycles) > 0,
		Count:  len(cycles),
		Cycles: cycles,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize cycles: %w", err)
	}

	cmd.Println(string(data))

	if len(cycles) > 0 {
		os.Exit(1)
	}
	return nil
}

func outputCyclesText(cmd *cobra.Command, cycles [][]string, g *graph.Graph, db *database.Database) error {
	width := 80

	cmd.Println(output.Header("Circular Dependency Analysis", width))
	cmd.Println()

	// Statistics section
	stats := g.Statistics()
	edges := 0
	if v, ok := stats["edges"].(int); ok {
		edges = v
	}
	cmd.Println("RTM Statistics:")
	cmd.Printf("  Total requirements: %d\n", db.Len())
	cmd.Printf("  Total dependencies: %d\n", edges)
	if db.Len() > 0 {
		avgDeps := float64(edges) / float64(db.Len())
		cmd.Printf("  Average dependencies per requirement: %.2f\n", avgDeps)
	}
	cmd.Println()

	if len(cycles) == 0 {
		cmd.Printf("%s No circular dependencies found!\n", output.Color("✓", output.Green))
		cmd.Println()
		cmd.Println("The dependency graph is acyclic (DAG).")
		return nil
	}

	cmd.Printf("%s FOUND %d CIRCULAR DEPENDENCY GROUP(S)\n\n",
		output.Color("✗", output.Red), len(cycles))

	// Summary section
	totalInvolved := 0
	largestCycle := 0
	smallestCycle := len(cycles[0])
	for _, cycle := range cycles {
		totalInvolved += len(cycle)
		if len(cycle) > largestCycle {
			largestCycle = len(cycle)
		}
		if len(cycle) < smallestCycle {
			smallestCycle = len(cycle)
		}
	}

	cmd.Println("Summary:")
	cmd.Printf("  Circular dependency groups: %d\n", len(cycles))
	cmd.Printf("  Requirements involved in cycles: %d\n", totalInvolved)
	cmd.Printf("  Largest cycle: %d requirements\n", largestCycle)
	cmd.Printf("  Smallest cycle: %d requirements\n", smallestCycle)
	cmd.Println()

	// Display cycles
	cmd.Println(output.SubHeader("TOP 10 LARGEST CIRCULAR DEPENDENCY GROUPS:", width))
	cmd.Println()

	limit := 10
	if len(cycles) < limit {
		limit = len(cycles)
	}

	for i := 0; i < limit; i++ {
		cycle := cycles[i]
		cmd.Printf("%d. Cycle with %d requirements:\n", i+1, len(cycle))

		// Find and display the path
		path := g.FindCyclePath(cycle)
		if len(path) > 0 {
			pathStr := strings.Join(path, " → ")
			cmd.Printf("   Path: %s\n", pathStr)
		} else {
			cmd.Printf("   Members: %s\n", strings.Join(cycle, ", "))
		}
		cmd.Println()
	}

	// Recommendations section
	cmd.Println(strings.Repeat("=", width))
	cmd.Println("RECOMMENDATIONS:")
	cmd.Println(strings.Repeat("=", width))
	cmd.Println()

	cmd.Println("1. Review dependency direction:")
	cmd.Println("   - Ensure parent requirements don't depend on child requirements")
	cmd.Println("   - Component requirements should depend on system requirements, not vice versa")
	cmd.Println()

	cmd.Println("2. Examine the largest cycles first:")
	cmd.Println("   - These likely indicate architectural issues")
	cmd.Println("   - May need to split into layers or stages")
	cmd.Println()

	cmd.Println("3. Check for \"blocks\" vs \"dependencies\" confusion:")
	cmd.Println("   - If A blocks B, then B depends on A (not both!)")
	cmd.Println("   - Run: rtmx reconcile --execute")
	cmd.Println()

	cmd.Println("4. Consider adding phase constraints:")
	cmd.Println("   - Requirements should not depend on later-phase requirements")
	cmd.Println()

	cmd.Printf("5. Total effort to fix:\n")
	cmd.Printf("   - %d requirements involved in %d cycles\n", totalInvolved, len(cycles))
	cmd.Println("   - Suggest reviewing in batches: largest cycles first")

	os.Exit(1)
	return nil
}

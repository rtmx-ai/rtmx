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

	cmd.Println(output.Header("Cycle Detection", width))
	cmd.Println()

	if len(cycles) == 0 {
		cmd.Printf("%s No circular dependencies found!\n", output.Color("✓", output.Green))
		cmd.Println()
		cmd.Println("The dependency graph is acyclic (DAG).")
		return nil
	}

	cmd.Printf("%s Found %d circular dependency chain(s)!\n\n",
		output.Color("!", output.Red), len(cycles))

	for i, cycle := range cycles {
		cmd.Println(output.SubHeader(fmt.Sprintf("Cycle %d (%d requirements)", i+1, len(cycle)), width))

		// Find and display the path
		path := g.FindCyclePath(cycle)
		if len(path) > 0 {
			cmd.Println()
			cmd.Println("  Cycle path:")
			for j, reqID := range path {
				req := db.Get(reqID)
				if req != nil {
					icon := output.StatusIcon(req.Status.String())
					arrow := ""
					if j < len(path)-1 {
						arrow = " → "
					}
					cmd.Printf("    %s %s%s\n", icon, reqID, arrow)
				}
			}
		} else {
			cmd.Println("  Members: " + strings.Join(cycle, ", "))
		}

		cmd.Println()

		// Show details for each member
		cmd.Println("  Details:")
		for _, reqID := range cycle {
			req := db.Get(reqID)
			if req != nil {
				deps := g.Dependencies(reqID)
				inCycleDeps := []string{}
				for _, d := range deps {
					for _, c := range cycle {
						if d == c {
							inCycleDeps = append(inCycleDeps, d)
							break
						}
					}
				}
				cmd.Printf("    %s depends on: %s\n", reqID, strings.Join(inCycleDeps, ", "))
			}
		}
		cmd.Println()
	}

	cmd.Println(output.Color("Resolution:", output.Bold))
	cmd.Println("  To break a cycle, remove or modify dependencies so that")
	cmd.Println("  at least one requirement in each cycle no longer depends")
	cmd.Println("  on another member of that cycle.")
	cmd.Println()
	cmd.Println("  Use 'rtmx reconcile' to help fix dependency issues.")

	os.Exit(1)
	return nil
}

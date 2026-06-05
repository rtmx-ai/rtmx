package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var (
	hygieneJSON      bool
	hygieneStrict    bool
	hygieneMinEffort float64
	hygieneMaxEffort float64
)

var hygieneCmd = &cobra.Command{
	Use:     "hygiene",
	Aliases: []string{"hygeine"},
	Short:   "Report RTM hygiene and actionability issues",
	Long: `Report RTM hygiene and actionability issues that can make requirements
hard to implement, verify, or assign.

The command is non-blocking by default and exits 0 after reporting findings.
Use --strict to return exit code 1 when any hygiene finding is present.`,
	RunE: runHygiene,
}

type HygieneFinding struct {
	Check   string `json:"check"`
	ReqID   string `json:"req_id,omitempty"`
	Message string `json:"message"`
}

type HygieneResult struct {
	Total    int              `json:"total"`
	Findings []HygieneFinding `json:"findings"`
	Summary  map[string]int   `json:"summary"`
	Cycles   [][]string       `json:"cycles,omitempty"`
}

func init() {
	hygieneCmd.Flags().BoolVar(&hygieneJSON, "json", false, "output as JSON")
	hygieneCmd.Flags().BoolVar(&hygieneStrict, "strict", false, "exit non-zero when findings are present")
	hygieneCmd.Flags().Float64Var(&hygieneMinEffort, "min-effort", 0.25, "minimum actionable effort in weeks")
	hygieneCmd.Flags().Float64Var(&hygieneMaxEffort, "max-effort", 0.5, "maximum actionable effort in weeks")
	rootCmd.AddCommand(hygieneCmd)
}

func runHygiene(cmd *cobra.Command, args []string) error {
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

	db, err := database.Load(cfg.DatabasePath(cwd))
	if err != nil {
		return fmt.Errorf("failed to load database: %w", err)
	}

	result := runHygieneChecks(db, hygieneMinEffort, hygieneMaxEffort)
	if hygieneJSON {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to serialize hygiene result: %w", err)
		}
		cmd.Println(string(data))
	} else {
		outputHygieneText(cmd, result)
	}

	if hygieneStrict && len(result.Findings) > 0 {
		return NewExitError(1, "hygiene findings present")
	}
	return nil
}

func runHygieneChecks(db *database.Database, minEffort, maxEffort float64) *HygieneResult {
	result := &HygieneResult{
		Total:   db.Len(),
		Summary: make(map[string]int),
	}

	for _, req := range db.All() {
		if req.EffortWeeks < minEffort || req.EffortWeeks > maxEffort {
			addHygieneFinding(result, "effort_bounds", req.ReqID,
				fmt.Sprintf("effort_weeks %.2f outside actionable range %.2f-%.2f", req.EffortWeeks, minEffort, maxEffort))
		}
		if strings.TrimSpace(req.Assignee) == "" || strings.EqualFold(strings.TrimSpace(req.Assignee), "team") {
			addHygieneFinding(result, "generic_owner", req.ReqID, "assignee is blank or generic")
		}
		if !req.HasTest() {
			addHygieneFinding(result, "missing_test_mapping", req.ReqID, "test_module and test_function are not both set")
		}
		if strings.TrimSpace(req.ExternalID) == "" {
			addHygieneFinding(result, "missing_external_id", req.ReqID, "external_id is blank")
		}
		if hasGenericAcceptanceCriteria(req) {
			addHygieneFinding(result, "generic_acceptance_criteria", req.ReqID, "requirement file contains generic acceptance criteria")
		}
	}

	g := graph.NewGraph(db)
	result.Cycles = g.FindCycles()
	for _, cycle := range result.Cycles {
		addHygieneFinding(result, "dependency_cycle", "", strings.Join(cycle, " -> "))
	}

	return result
}

func addHygieneFinding(result *HygieneResult, check, reqID, message string) {
	result.Findings = append(result.Findings, HygieneFinding{
		Check:   check,
		ReqID:   reqID,
		Message: message,
	})
	result.Summary[check]++
}

func hasGenericAcceptanceCriteria(req *database.Requirement) bool {
	if strings.TrimSpace(req.RequirementFile) == "" {
		return false
	}
	data, err := os.ReadFile(req.RequirementFile)
	if err != nil {
		return false
	}
	text := string(data)
	genericPhrases := []string{
		"Requirement language is reviewed",
		"Validation evidence is implemented or explicitly planned",
		"Evidence can be mapped back to this requirement",
	}
	for _, phrase := range genericPhrases {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return false
}

func outputHygieneText(cmd *cobra.Command, result *HygieneResult) {
	cmd.Println(output.Header("RTM Hygiene Check", 80))
	cmd.Println()
	cmd.Printf("Requirements: %d\n", result.Total)
	cmd.Printf("Findings:     %d\n", len(result.Findings))
	cmd.Println()

	if len(result.Summary) > 0 {
		cmd.Println(output.SubHeader("Summary", 80))
		for check, count := range result.Summary {
			cmd.Printf("  %s: %d\n", check, count)
		}
		cmd.Println()
	}

	if len(result.Findings) == 0 {
		cmd.Printf("%s No hygiene findings.\n", output.Color("✓", output.Green))
		return
	}

	cmd.Println(output.SubHeader("Findings", 80))
	limit := len(result.Findings)
	if limit > 50 {
		limit = 50
	}
	for i := 0; i < limit; i++ {
		finding := result.Findings[i]
		if finding.ReqID != "" {
			cmd.Printf("- [%s] %s: %s\n", finding.Check, finding.ReqID, finding.Message)
		} else {
			cmd.Printf("- [%s] %s\n", finding.Check, finding.Message)
		}
	}
	if len(result.Findings) > limit {
		cmd.Printf("... %d more findings\n", len(result.Findings)-limit)
	}
}

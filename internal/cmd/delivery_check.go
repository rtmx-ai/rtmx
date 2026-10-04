package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/orchestration/deliverycheck"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var (
	deliveryCheckReq    string
	deliveryCheckBase   string
	deliveryCheckStrict bool
	deliveryCheckJSON   bool
	deliveryCheckTitle  string
	deliveryCheckBody   string
)

var deliveryCheckCmd = &cobra.Command{
	Use:   "delivery-check",
	Short: "Warn-first check for one PR per REQ and one commit per AC",
	Long: `Report whether the current branch maps to exactly one requirement and
whether each commit since the merge-base names that requirement and an
acceptance-criterion marker (AC1, AC-1, or ac_id=…).

Default mode prints warnings and exits 0. Use --strict to exit 1 on
violations. Prefer this over inventing commit/PR process in the agent.`,
	RunE: runDeliveryCheck,
}

func init() {
	deliveryCheckCmd.Flags().StringVar(&deliveryCheckReq, "req", "", "force requirement ID")
	deliveryCheckCmd.Flags().StringVar(&deliveryCheckBase, "base", "main", "merge-base ref")
	deliveryCheckCmd.Flags().BoolVar(&deliveryCheckStrict, "strict", false, "exit 1 when checks fail")
	deliveryCheckCmd.Flags().BoolVar(&deliveryCheckJSON, "json", false, "output as JSON")
	deliveryCheckCmd.Flags().StringVar(&deliveryCheckTitle, "title", "", "PR title override (tests / offline)")
	deliveryCheckCmd.Flags().StringVar(&deliveryCheckBody, "body", "", "PR body override (tests / offline)")
	rootCmd.AddCommand(deliveryCheckCmd)
}

func runDeliveryCheck(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	title := deliveryCheckTitle
	body := deliveryCheckBody
	if title == "" && body == "" {
		title, body = lookupPRMeta(cwd)
	}
	res, err := deliverycheck.Check(deliverycheck.Input{
		Dir:   cwd,
		Base:  deliveryCheckBase,
		ReqID: deliveryCheckReq,
		Title: title,
		Body:  body,
	})
	if err != nil {
		return err
	}
	if deliveryCheckJSON {
		data, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			return err
		}
		cmd.Println(string(data))
	} else {
		printDeliveryCheck(cmd, res)
	}
	if deliveryCheckStrict && !res.OK {
		return NewExitError(1, "delivery-check failed")
	}
	return nil
}

func printDeliveryCheck(cmd *cobra.Command, res *deliverycheck.Result) {
	if res.OK {
		cmd.Printf("%s delivery-check ok", output.Color("✓", output.Green))
		if res.ReqID != "" {
			cmd.Printf(" (%s)", res.ReqID)
		}
		cmd.Println()
		cmd.Printf("  %d commit(s) mapped\n", len(res.Commits))
		return
	}
	cmd.Printf("%s delivery-check warnings\n", output.Color("!", output.Yellow))
	if res.ReqID != "" {
		cmd.Printf("  req_id: %s\n", res.ReqID)
	}
	for _, w := range res.Warnings {
		cmd.Printf("  - %s\n", w)
	}
}

func lookupPRMeta(cwd string) (string, string) {
	c := exec.Command("gh", "pr", "view", "--json", "title,body")
	c.Dir = cwd
	out, err := c.Output()
	if err != nil {
		return "", ""
	}
	var row struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.Unmarshal(out, &row); err != nil {
		return "", ""
	}
	return row.Title, row.Body
}

// reportDeliveryCheckForGate prints a non-blocking delivery-check section for release gate.
func reportDeliveryCheckForGate(cmd *cobra.Command, cwd string) {
	res, err := deliverycheck.Check(deliverycheck.Input{Dir: cwd, Base: "main"})
	if err != nil {
		cmd.Printf("  %s delivery-check skipped: %v\n", output.Color("WARN", output.Yellow), err)
		return
	}
	if res.OK {
		cmd.Printf("  %s delivery-check: ok (%d commits)\n", output.Color("PASS", output.Green), len(res.Commits))
		return
	}
	cmd.Printf("  %s delivery-check (non-blocking): %d issue(s)\n", output.Color("WARN", output.Yellow), len(res.Errors))
	for _, e := range res.Errors {
		if len(e) > 80 {
			e = e[:80] + "…"
		}
		cmd.Printf("    - %s\n", e)
	}
	_ = strings.TrimSpace
}

func formatDeliveryCheckSummary(res *deliverycheck.Result) string {
	if res.OK {
		return fmt.Sprintf("ok (%d commits)", len(res.Commits))
	}
	return fmt.Sprintf("%d issue(s)", len(res.Errors))
}

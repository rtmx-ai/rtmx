package cmd

import (
	"encoding/json"
	"os"

	"github.com/rtmx-ai/rtmx/internal/orchestration/trade"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var (
	tradeOpenReq   string
	tradeOpenTitle string
	tradeListReq   string
	tradeJSON      bool
	tradeChoice    string
)

var tradeCmd = &cobra.Command{
	Use:   "trade",
	Short: "Trade-analysis checkpoints under .rtmx/trades/",
	Long: `Open, list, and resolve trade-analysis checkpoints.

Trades capture multi-option intent before coding. loop_tick sets
trade_required while an open trade exists for the selected requirement.`,
}

var tradeOpenCmd = &cobra.Command{
	Use:   "open",
	Short: "Open a trade analysis for a requirement",
	RunE:  runTradeOpen,
}

var tradeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List trade analyses",
	RunE:  runTradeList,
}

var tradeResolveCmd = &cobra.Command{
	Use:   "resolve TRADE-ID",
	Short: "Resolve a trade with a chosen option",
	Args:  cobra.ExactArgs(1),
	RunE:  runTradeResolve,
}

func init() {
	tradeOpenCmd.Flags().StringVar(&tradeOpenReq, "req", "", "requirement ID")
	tradeOpenCmd.Flags().StringVar(&tradeOpenTitle, "title", "", "trade title")
	_ = tradeOpenCmd.MarkFlagRequired("req")
	tradeListCmd.Flags().StringVar(&tradeListReq, "req", "", "filter by requirement ID")
	tradeListCmd.Flags().BoolVar(&tradeJSON, "json", false, "output as JSON")
	tradeResolveCmd.Flags().StringVar(&tradeChoice, "choice", "", "chosen option")
	_ = tradeResolveCmd.MarkFlagRequired("choice")
	tradeResolveCmd.Flags().BoolVar(&tradeJSON, "json", false, "output as JSON")
	tradeOpenCmd.Flags().BoolVar(&tradeJSON, "json", false, "output as JSON")

	tradeCmd.AddCommand(tradeOpenCmd)
	tradeCmd.AddCommand(tradeListCmd)
	tradeCmd.AddCommand(tradeResolveCmd)
	rootCmd.AddCommand(tradeCmd)
}

func runTradeOpen(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	t, err := trade.Open(cwd, tradeOpenReq, tradeOpenTitle)
	if err != nil {
		return err
	}
	if tradeJSON {
		data, _ := json.MarshalIndent(t, "", "  ")
		cmd.Println(string(data))
		return nil
	}
	cmd.Printf("%s Opened %s\n", output.Color("✓", output.Green), t.ID)
	cmd.Printf("  %s\n", t.Path)
	return nil
}

func runTradeList(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	trades, err := trade.List(cwd, tradeListReq)
	if err != nil {
		return err
	}
	if tradeJSON {
		data, _ := json.MarshalIndent(trades, "", "  ")
		cmd.Println(string(data))
		return nil
	}
	if len(trades) == 0 {
		cmd.Println("No trades.")
		return nil
	}
	for _, t := range trades {
		cmd.Printf("%-40s  %-10s  %s\n", t.ID, t.Status, t.Title)
	}
	return nil
}

func runTradeResolve(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	t, err := trade.Resolve(cwd, args[0], tradeChoice)
	if err != nil {
		return err
	}
	if tradeJSON {
		data, _ := json.MarshalIndent(t, "", "  ")
		cmd.Println(string(data))
		return nil
	}
	cmd.Printf("%s Resolved %s → %s\n", output.Color("✓", output.Green), t.ID, tradeChoice)
	return nil
}

func tradeRequiredFor(cwd, reqID, reqText, mdBody string) bool {
	open, err := trade.HasOpen(cwd, reqID)
	if err == nil && open {
		return true
	}
	return trade.NeedsTrade(reqText, mdBody)
}

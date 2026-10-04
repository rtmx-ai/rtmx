package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/orchestration/trade"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func newTradeTestCmd() *cobra.Command {
	root := &cobra.Command{Use: "rtmx", SilenceUsage: true, SilenceErrors: true}
	var req, title, choice string
	var jsonOut bool
	tr := &cobra.Command{Use: "trade"}
	open := &cobra.Command{
		Use: "open",
		RunE: func(cmd *cobra.Command, args []string) error {
			tradeOpenReq = req
			tradeOpenTitle = title
			tradeJSON = jsonOut
			return runTradeOpen(cmd, args)
		},
	}
	open.Flags().StringVar(&req, "req", "", "")
	open.Flags().StringVar(&title, "title", "", "")
	open.Flags().BoolVar(&jsonOut, "json", false, "")
	list := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			tradeListReq = req
			tradeJSON = jsonOut
			return runTradeList(cmd, args)
		},
	}
	list.Flags().StringVar(&req, "req", "", "")
	list.Flags().BoolVar(&jsonOut, "json", false, "")
	resolve := &cobra.Command{
		Use:  "resolve",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tradeChoice = choice
			tradeJSON = jsonOut
			return runTradeResolve(cmd, args)
		},
	}
	resolve.Flags().StringVar(&choice, "choice", "", "")
	resolve.Flags().BoolVar(&jsonOut, "json", false, "")
	tr.AddCommand(open, list, resolve)
	root.AddCommand(tr)
	return root
}

func TestTradeCheckpoints(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-023d")

	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	buf := new(bytes.Buffer)
	cmd := newTradeTestCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"trade", "open", "--req", "REQ-EX-050", "--title", "Pick store", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("open: %v\n%s", err, buf.String())
	}
	var opened trade.Trade
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &opened); err != nil {
		t.Fatalf("json: %v\n%s", err, buf.String())
	}
	if opened.Status != trade.StatusOpen || opened.ReqID != "REQ-EX-050" {
		t.Fatalf("opened: %+v", opened)
	}
	if _, err := os.Stat(filepath.Join(dir, ".rtmx", "trades", opened.ID+".md")); err != nil {
		t.Fatal(err)
	}

	if !tradeRequiredFor(dir, "REQ-EX-050", "do something", "") {
		t.Fatal("expected trade_required with open trade")
	}

	buf.Reset()
	cmd = newTradeTestCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"trade", "resolve", opened.ID, "--choice", "Option A", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("resolve: %v\n%s", err, buf.String())
	}
	open, err := trade.HasOpen(dir, "REQ-EX-050")
	if err != nil || open {
		t.Fatalf("expected no open trade: %v %v", open, err)
	}

	if !trade.NeedsTrade("decide TBD", "") {
		t.Fatal("TBD should need trade")
	}

	// loop plan flags trade_required
	db := testDBHeader +
		"REQ-EX-050,CLI,Commands,Ambiguous either/or choice,Pass,mod,TestA,Unit Test,MISSING,HIGH,1,,0.5,,,,,,,\n"
	proj := setupNextTestProject(t, db)
	writeLoopReqMD(t, proj, "REQ-EX-050", 2)
	if _, err := trade.Open(proj, "REQ-EX-050", "Ambiguity"); err != nil {
		t.Fatal(err)
	}
	openPRSource = staticOpenPRs{}
	t.Cleanup(func() { openPRSource = nil })
	if err := os.Chdir(proj); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	tick := newLoopTickCLICmd()
	tick.SetOut(buf)
	tick.SetErr(buf)
	tick.SetArgs([]string{"loop", "tick", "--agent-id", "trader"})
	if err := tick.Execute(); err != nil {
		t.Fatalf("tick: %v\n%s", err, buf.String())
	}
	var plan loopPlan
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &plan); err != nil {
		t.Fatalf("plan: %v\n%s", err, buf.String())
	}
	if !plan.TradeRequired || plan.ReadyToImplement {
		t.Fatalf("want trade_required and not ready: %+v", plan)
	}
}

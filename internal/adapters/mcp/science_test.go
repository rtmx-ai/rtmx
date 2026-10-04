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

func TestScientificMCPTools(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-023c")

	dir := t.TempDir()
	rtmxDir := filepath.Join(dir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	csv := `req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id
REQ-SCI-001,CLI,Commands,Atomic science work,Pass,mod,TestA,Unit Test,MISSING,HIGH,1,,0.5,,,,,,,
`
	dbPath := filepath.Join(rtmxDir, "database.csv")
	if err := os.WriteFile(dbPath, []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	mdDir := filepath.Join(rtmxDir, "requirements", "CLI")
	if err := os.MkdirAll(mdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	md := "# REQ-SCI-001\n\n## Acceptance Criteria\n\n1. one\n2. two\n"
	if err := os.WriteFile(filepath.Join(mdDir, "REQ-SCI-001.md"), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := NewServer(dbPath, config.DefaultConfig(), WithQuiet(true))
	load := func(t *testing.T) *database.Database {
		t.Helper()
		db, err := database.Load(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		return db
	}

	list := srv.handleToolsList().(map[string]interface{})
	tools := list["tools"].([]toolDef)
	names := map[string]bool{}
	for _, tool := range tools {
		names[tool.Name] = true
	}
	for _, want := range []string{
		"loop_tick", "decompose", "hygiene", "cycles", "webs", "context", "delivery_check",
		"trade_open", "trade_list", "trade_resolve",
	} {
		if !names[want] {
			t.Errorf("missing tool %s", want)
		}
	}
	if names["run_cli"] {
		t.Error("run_cli must not be exposed")
	}

	if srv.toolHygiene(load(t), map[string]interface{}{}) == nil {
		t.Fatal("hygiene nil")
	}
	cyc := srv.toolCycles(load(t)).(map[string]interface{})
	if cyc["found"] != false {
		t.Fatalf("cycles: %+v", cyc)
	}
	webs := srv.toolWebs(load(t), map[string]interface{}{}).(map[string]interface{})
	if webs["count"].(int) < 1 {
		t.Fatalf("webs: %+v", webs)
	}
	ctx := srv.toolContext(load(t)).(map[string]interface{})
	if ctx["total"].(int) != 1 {
		t.Fatalf("context: %+v", ctx)
	}

	planRaw, rpcErr := srv.toolLoopTick(load(t), map[string]interface{}{"agent_id": "sci-agent"})
	if rpcErr != nil {
		t.Fatalf("loop_tick rpc: %+v", rpcErr)
	}
	planBytes, _ := json.Marshal(planRaw)
	var plan scienceLoopPlan
	if err := json.Unmarshal(planBytes, &plan); err != nil {
		t.Fatalf("plan: %v %s", err, planBytes)
	}
	if plan.Idle || plan.ReqID != "REQ-SCI-001" || !plan.Claimed {
		t.Fatalf("plan: %+v", plan)
	}

	dec, rpcErr := srv.toolDecompose(load(t), map[string]interface{}{
		"req_id": "REQ-SCI-001", "agent_id": "sci-agent",
	})
	if rpcErr != nil {
		t.Fatalf("decompose: %+v", rpcErr)
	}
	if dec.(map[string]interface{})["atomic"] != true {
		t.Fatalf("expected atomic: %+v", dec)
	}

	if _, rpcErr := srv.toolDeliveryCheck(map[string]interface{}{"req_id": "REQ-SCI-001"}); rpcErr != nil {
		t.Fatalf("delivery_check rpc: %+v", rpcErr)
	}

	opened, rpcErr := srv.toolTradeOpen(map[string]interface{}{
		"req_id": "REQ-SCI-001", "title": "Choice", "agent_id": "sci-agent",
	})
	if rpcErr != nil {
		t.Fatalf("trade_open: %+v", rpcErr)
	}
	tradeID, _ := opened.(map[string]interface{})["id"].(string)
	listed, rpcErr := srv.toolTradeList(map[string]interface{}{"req_id": "REQ-SCI-001"})
	if rpcErr != nil {
		t.Fatalf("trade_list: %+v", rpcErr)
	}
	if listed.(map[string]interface{})["count"].(int) != 1 {
		t.Fatalf("trade_list: %+v", listed)
	}
	if _, rpcErr := srv.toolTradeResolve(map[string]interface{}{
		"trade_id": tradeID, "choice": "A", "agent_id": "sci-agent",
	}); rpcErr != nil {
		t.Fatalf("trade_resolve: %+v", rpcErr)
	}
}

package output

import (
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestTUIFormat(t *testing.T) {
	rtmx.Req(t, "REQ-PLUGIN-004")

	t.Run("status_compact", func(t *testing.T) {
		data := TUIStatusData{
			Complete:   138,
			Partial:    12,
			Missing:    18,
			Total:      168,
			Percentage: 82.0,
			NextReqID:  "REQ-SYNC-003",
			NextPrio:   "P0",
			NextEffort: "2wk",
		}

		out := RenderTUIStatus(data)
		lines := strings.Split(strings.TrimSpace(out), "\n")

		if len(lines) > 8 {
			t.Errorf("TUI status must be <= 8 lines, got %d", len(lines))
		}
		if !strings.Contains(out, "82%") {
			t.Error("status should show completion percentage")
		}
		if !strings.Contains(out, "138/168") {
			t.Error("status should show complete/total ratio")
		}
		if !strings.Contains(out, "COMPLETE 138") {
			t.Error("status should show COMPLETE count")
		}
		if !strings.Contains(out, "REQ-SYNC-003") {
			t.Error("status should show next unblocked requirement")
		}
		if strings.Contains(out, "\033[") {
			t.Error("TUI output must not contain ANSI escape codes")
		}
	})

	t.Run("status_no_next", func(t *testing.T) {
		data := TUIStatusData{
			Complete:   168,
			Partial:    0,
			Missing:    0,
			Total:      168,
			Percentage: 100.0,
		}

		out := RenderTUIStatus(data)
		if strings.Contains(out, "Next unblocked") {
			t.Error("should not show next line when no open requirements")
		}
	})

	t.Run("backlog_aligned", func(t *testing.T) {
		items := []TUIBacklogItem{
			{1, "REQ-SYNC-003", "P0", "2wk", "SYNC", "MISSING"},
			{2, "REQ-MCP-003", "P0", "1wk", "MCP", "MISSING"},
			{3, "REQ-AGENT-001", "P0", "1wk", "AGENT", "MISSING"},
		}

		out := RenderTUIBacklog(items, 18)
		lines := strings.Split(strings.TrimSpace(out), "\n")

		if len(lines) > 8 {
			t.Errorf("TUI backlog must be <= 8 lines, got %d", len(lines))
		}
		if !strings.Contains(lines[0], "18 remaining") {
			t.Error("backlog should show total open count")
		}
		if !strings.Contains(out, "REQ-SYNC-003") {
			t.Error("backlog should contain items")
		}
		if strings.Contains(out, "\033[") {
			t.Error("TUI output must not contain ANSI escape codes")
		}
	})

	t.Run("health_compact", func(t *testing.T) {
		checks := []TUIHealthCheck{
			{"coverage", "82% (threshold 80%)", "PASS"},
			{"consistency", "1 issue", "WARN"},
			{"dependencies", "no cycles", "PASS"},
		}

		out := RenderTUIHealth("WARNING", checks)
		lines := strings.Split(strings.TrimSpace(out), "\n")

		if len(lines) > 8 {
			t.Errorf("TUI health must be <= 8 lines, got %d", len(lines))
		}
		if !strings.Contains(lines[0], "WARNING") {
			t.Error("health should show overall status")
		}
		if !strings.Contains(out, "coverage") {
			t.Error("health should list checks")
		}
		if strings.Contains(out, "\033[") {
			t.Error("TUI output must not contain ANSI escape codes")
		}
	})
}

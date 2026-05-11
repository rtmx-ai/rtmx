package output

import (
	"fmt"
	"strings"
)

// TUIStatusData holds the data for TUI status rendering.
type TUIStatusData struct {
	Complete   int
	Partial    int
	Missing    int
	Total      int
	Percentage float64
	NextReqID  string
	NextPrio   string
	NextEffort string
}

// RenderTUIStatus produces compact, markdown-safe status output (3 lines max).
func RenderTUIStatus(d TUIStatusData) string {
	var b strings.Builder
	fmt.Fprintf(&b, "RTMX  %.0f%% complete  %d/%d reqs\n", d.Percentage, d.Complete, d.Total)
	fmt.Fprintf(&b, "COMPLETE %d | PARTIAL %d | MISSING %d\n", d.Complete, d.Partial, d.Missing)
	if d.NextReqID != "" {
		fmt.Fprintf(&b, "Next unblocked: %s (%s, %s)", d.NextReqID, d.NextPrio, d.NextEffort)
	}
	return b.String()
}

// TUIBacklogItem holds one backlog entry for TUI rendering.
type TUIBacklogItem struct {
	Rank     int
	ReqID    string
	Priority string
	Effort   string
	Category string
	Status   string
}

// RenderTUIBacklog produces compact, aligned backlog output.
func RenderTUIBacklog(items []TUIBacklogItem, totalOpen int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Open Requirements (%d remaining)\n", totalOpen)
	for _, item := range items {
		fmt.Fprintf(&b, "%d. %-14s %-5s %-4s %-8s %s\n",
			item.Rank, item.ReqID, item.Priority, item.Effort, item.Category, item.Status)
	}
	return b.String()
}

// TUIHealthCheck holds one health check result for TUI rendering.
type TUIHealthCheck struct {
	Name    string
	Detail  string
	Status  string // PASS, WARN, FAIL
}

// RenderTUIHealth produces compact health output.
func RenderTUIHealth(overall string, checks []TUIHealthCheck) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Health: %s\n", overall)
	for _, c := range checks {
		fmt.Fprintf(&b, "  %s: %s  %s\n", c.Name, c.Detail, c.Status)
	}
	return b.String()
}

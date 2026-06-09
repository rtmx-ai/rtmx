# REQ-DASH-044: Enrich Graph Node JSON with Full Requirement Data

## Summary

The graph JSON payload served by the Go backend currently includes only 3 fields per node (id, status, group). This requirement enriches each node with effort_weeks, priority, phase, assignee, sprint, is_blocked, is_critical, and is_cycle_member -- all derived from the existing Requirement model and graph analysis functions. This is a backend-only change to the gNode struct and the graph partial handler in `serve_dashboard.go`. The frontend is not modified; it simply receives richer data for future visual encoding requirements.

## Acceptance Criteria

1. The `gNode` JSON struct includes these fields with correct types:
   - `id` (string) -- requirement ID (existing)
   - `status` (string) -- requirement status (existing)
   - `group` (string) -- category (existing)
   - `effort_weeks` (float64) -- from `Requirement.EffortWeeks`
   - `priority` (string) -- from `Requirement.Priority` (P0/P1/P2/P3)
   - `phase` (int) -- from `Requirement.Phase`
   - `assignee` (string) -- from `Requirement.Assignee`
   - `sprint` (string) -- from `Requirement.Sprint`
   - `is_blocked` (bool) -- from `graph.IsBlocked(reqID)`
   - `is_critical` (bool) -- true if reqID appears in `graph.CriticalPath()`
   - `is_cycle_member` (bool) -- true if reqID appears in any cycle from `graph.FindCycles()`
2. CriticalPath() and FindCycles() are called once per request, not once per node.
3. The JSON payload size increase is proportional (no duplication of bulk text fields like requirement_text or notes).
4. Existing frontend `renderGraph()` continues to work because new fields are additive.
5. The `/api/graph` JSON endpoint (if it exists) returns the same enriched structure.

## Dependencies

- REQ-DASH-001 (SPA framework)
- REQ-DASH-010 (design tokens)

## Blocks

- REQ-DASH-050+ (visual encoding requirements that use effort, priority, blocked state for node sizing, coloring, opacity)

## Files to Modify

- `internal/cmd/serve_dashboard.go` (expand gNode struct, populate new fields in graph handler)

## Test Strategy

- Unit test: mock database with requirements of varying effort/priority/assignee, verify JSON output contains all expected fields with correct values
- Unit test: verify is_critical=true only for nodes on the critical path
- Unit test: verify is_cycle_member=true only for nodes in cycles
- Unit test: verify is_blocked=true only for nodes with incomplete dependencies
- Integration test: load graph partial, parse embedded JSON, verify all fields present
- Performance test: enriched payload for 200-node graph completes in under 50ms

## Effort

- 0.75 weeks

## Priority

- HIGH

## Phase

- 32

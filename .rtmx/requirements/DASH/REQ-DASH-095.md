# REQ-DASH-095: Add blocking_ids Array to Enriched Node JSON

## Summary

Each node in the enriched graph JSON includes a `blocking_ids` field: an array of strings listing the IDs of incomplete dependencies that block this node. This enables the tooltip (REQ-DASH-054) to show which specific requirements are blocking, and the highlight mode (REQ-DASH-062) to trace blocking chains. The field is computed server-side by checking each dependency's status.

## Acceptance Criteria

1. Each node object in the graph JSON response includes a `blocking_ids` field of type `[]string`.
2. `blocking_ids` contains the IDs of direct dependencies that have status other than COMPLETE.
3. If all dependencies are COMPLETE (or the node has no dependencies), `blocking_ids` is an empty array `[]`, not null.
4. The field is computed in the graph enrichment handler in serve_dashboard.go.
5. The JSON schema change is backward-compatible: existing consumers ignore unknown fields.
6. Unit tests verify correct `blocking_ids` for nodes with 0, 1, and multiple incomplete dependencies.

## Dependencies

- REQ-DASH-044 (enriched graph JSON structure to extend)

## Blocks

- REQ-DASH-054 (blocked node tooltip needs blocking_ids to display blocker details)
- REQ-DASH-062 (highlight mode needs blocking_ids to trace chains)

## Files to Modify

- `internal/cmd/serve_dashboard.go` (add blocking_ids computation to graph enrichment)

## Test Strategy

- Unit test: node with no dependencies has empty blocking_ids array
- Unit test: node with all COMPLETE dependencies has empty blocking_ids array
- Unit test: node with one incomplete dependency has that ID in blocking_ids
- Unit test: node with mixed complete/incomplete dependencies lists only incomplete IDs
- Integration test: GET /api/graph response includes blocking_ids on all nodes

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 32

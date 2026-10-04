# REQ-DASH-047: Add Web (Connected Component) Data to Graph JSON

## Summary

Each node in the graph JSON payload shall include a `web_id` field (integer) identifying which connected component (web) the node belongs to, derived from `graph.DetectWebs()`. Complete requirements not assigned to any web receive `web_id: -1`. This enables the frontend to visually group or color-code independent work streams, making parallelizable work immediately visible.

## Acceptance Criteria

1. Each `gNode` JSON object includes a `web_id` field (int).
2. The web_id is a stable zero-based index corresponding to the web's position in the `DetectWebs()` result (sorted by total effort descending, as DetectWebs() already does).
3. Nodes that appear in the same web share the same web_id.
4. Complete requirements (excluded from DetectWebs()) receive `web_id: -1`.
5. The graph JSON also includes a top-level `web_count` field (int) indicating the number of detected webs.
6. When all requirements are complete, web_count is 0 and all nodes have `web_id: -1`.

## Dependencies

- REQ-DASH-001 (SPA framework)
- REQ-DASH-010 (design tokens)

## Blocks

- REQ-DASH-050+ (visual grouping/coloring by web)

## Files to Modify

- `internal/cmd/serve_dashboard.go` (add web_id to gNode, add web_count to gJSON, call DetectWebs() and build ID-to-web-index lookup)

## Test Strategy

- Unit test: two disconnected subgraphs produce two distinct web_ids
- Unit test: single connected subgraph produces one web_id for all incomplete nodes
- Unit test: complete requirements get web_id -1
- Unit test: web_count matches len(DetectWebs())
- Unit test: web_id indices are contiguous starting from 0

## Effort

- 0.50 weeks

## Priority

- MEDIUM

## Phase

- 32

# REQ-DASH-046: Add Layer Data to Graph JSON from graph.Layers()

## Summary

Each node in the graph JSON payload shall include a `layer` field (integer) representing its depth from root nodes, computed by `graph.Layers()`. Layer 0 contains root nodes (no dependencies), layer 1 contains nodes whose dependencies are all in layer 0, and so on. This layering drives the vertical (or horizontal) positioning axis in the dagre layout, ensuring the visual hierarchy reflects the actual dependency depth.

## Acceptance Criteria

1. Each `gNode` JSON object includes a `layer` field (int).
2. The layer value is derived from `graph.Layers()`, which returns `[][]string` (each index is a layer, containing requirement IDs at that depth).
3. Root nodes (no dependencies) have `layer: 0`.
4. A node's layer is always strictly greater than the layer of every node it depends on.
5. Nodes involved in cycles (which Layers() may not fully resolve) receive a layer value of -1 to indicate indeterminate depth.
6. The graph JSON also includes a top-level `layer_count` field (int) indicating the total number of layers.
7. When filtering by category, layers are recomputed for the filtered subgraph (not the full graph).

## Dependencies

- REQ-DASH-001 (SPA framework)
- REQ-DASH-010 (design tokens)

## Blocks

- REQ-DASH-048 (dagre layout uses layer data for vertical positioning)

## Files to Modify

- `internal/cmd/serve_dashboard.go` (add layer field to gNode, add layer_count to gJSON, call Layers() and build lookup map)

## Test Strategy

- Unit test: linear chain A->B->C produces layers [A:0, B:1, C:2]
- Unit test: diamond A->B, A->C, B->D, C->D produces layers [A:0, B:1, C:1, D:2]
- Unit test: disconnected graph has multiple roots at layer 0
- Unit test: cycle members receive layer -1
- Unit test: category-filtered graph computes layers only for visible nodes
- Unit test: verify layer_count equals the number of distinct layers

## Effort

- 0.50 weeks

## Priority

- HIGH

## Phase

- 32

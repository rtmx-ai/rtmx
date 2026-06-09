# REQ-DASH-062: Dependency Chain Highlight Mode

## Summary

Clicking a graph node shall activate highlight mode, visually tracing the full
transitive dependency chain. Upstream dependencies (nodes this requirement depends on)
are tinted blue, downstream dependents (nodes that depend on this requirement) are
tinted amber, and the selected node is highlighted with an emerald glow. All other
nodes and edges dim to 0.3 opacity. Clicking the graph background clears the highlight.

## Acceptance Criteria

1. Clicking a node activates highlight mode for that node's dependency chain
2. The selected node receives an emerald glow effect (`--color-emerald-500` stroke, 3px width)
3. Upstream transitive dependencies (all ancestors via `blocked_by` edges) are tinted with `--color-blue-400` at 0.8 opacity
4. Downstream transitive dependents (all descendants via `blocks` edges) are tinted with `--color-amber-400` at 0.8 opacity
5. Edges within the highlighted chain retain full opacity and match the direction color (blue for upstream, amber for downstream)
6. All nodes and edges outside the chain dim to 0.3 opacity
7. Clicking the SVG background (not a node) clears highlight mode and restores all nodes to full opacity
8. Pressing Escape also clears highlight mode
9. Highlight mode coexists with the detail panel (REQ-DASH-058): clicking a node opens the panel AND highlights the chain
10. Transitive traversal correctly handles diamond dependencies (node appears in both upstream and downstream paths)

## Dependencies

- REQ-DASH-044 (enriched JSON provides dependency graph edges for transitive traversal)
- REQ-DASH-048 (dagre layout engine renders nodes and edges that can be styled)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/graph.js`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: clicking a node computes correct transitive upstream and downstream sets
- Unit test: upstream nodes receive blue tint, downstream nodes receive amber tint
- Unit test: non-chain nodes and edges dim to 0.3 opacity
- Unit test: clicking background clears all highlight styling
- Unit test: diamond dependency does not cause infinite loop or duplicate highlighting
- Integration test: highlight mode with real dependency data from enriched JSON endpoint

## Effort

- 0.75 weeks

## Priority

- MEDIUM

## Phase

- 32

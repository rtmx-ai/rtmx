# REQ-DASH-089: Cluster Threshold Control UI (075c)

## Summary

A dropdown control in the graph toolbar allows the user to select the clustering behavior: "Auto" (cluster when > 100 nodes, the default), "Always cluster" (cluster regardless of node count), or "Never cluster" (show all individual nodes). The selected mode is stored in Alpine.js state and persists within the session.

## Acceptance Criteria

1. A dropdown control appears in the graph toolbar with three options: "Auto", "Always cluster", "Never cluster".
2. "Auto" (default) clusters when node count exceeds 100.
3. "Always cluster" forces clustering regardless of node count.
4. "Never cluster" shows all individual nodes regardless of count.
5. Changing the dropdown re-renders the graph with the selected clustering mode.
6. The selected mode is stored in Alpine.js component state and survives htmx partial reloads.
7. The dropdown uses design token styling consistent with other graph controls.

## Dependencies

- REQ-DASH-075 (cluster rendering logic must exist to be controlled)
- REQ-DASH-010 (design tokens for dropdown styling)

## Blocks

- None currently identified

## Files to Modify

- `internal/dashboard/templates/partials/graph.html` (cluster threshold dropdown in toolbar)
- `internal/dashboard/static/app.js` (dropdown change handler, clustering mode state)

## Test Strategy

- Unit test: selecting "Always cluster" forces clustering on a 50-node graph
- Unit test: selecting "Never cluster" disables clustering on a 200-node graph
- Unit test: selecting "Auto" restores default threshold behavior
- Integration test: dropdown change triggers graph re-render with correct mode

## Effort

- 0.25 weeks

## Priority

- MEDIUM

## Phase

- 32

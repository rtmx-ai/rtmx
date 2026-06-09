# REQ-DASH-061: Multi-Select Filter Controls for Graph View

## Summary

The graph view shall provide multi-select filter controls for status, priority, and
category dimensions. Selecting filter values hides non-matching nodes and their
connected edges from the graph. Filter state persists across graph re-renders
(e.g., group-by changes, zoom resets) and is reflected in URL query parameters.

## Acceptance Criteria

1. Three multi-select dropdown controls appear in the graph toolbar: Status, Priority, Category
2. Each dropdown shows checkboxes for all available values in that dimension, derived from the current dataset
3. Selecting one or more values in a filter hides nodes that do not match ANY selected value in that dimension
4. Multiple filters combine with AND logic: a node must match at least one value in each active filter
5. Edges connected to hidden nodes are also hidden
6. Filter state persists in URL query parameters (`?status=COMPLETE,IN_PROGRESS&priority=HIGH`)
7. Filter state survives group-by changes (REQ-DASH-060) and zoom/fit operations (REQ-DASH-064)
8. A "Clear filters" button appears when any filter is active, resetting all filters
9. Active filter count is displayed as a badge on each dropdown (e.g., "Status (2)")
10. Dagre re-layout runs after filtering to eliminate gaps left by hidden nodes

## Dependencies

- REQ-DASH-048 (dagre layout engine renders filterable node elements)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/graph.js`
- `internal/dashboard/templates/partials/graph.html`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: selecting a status filter hides non-matching nodes and their edges
- Unit test: multiple filters combine with AND logic
- Unit test: "Clear filters" resets all dropdowns and shows all nodes
- Unit test: filter badge counts update when selections change
- Unit test: dagre re-layout removes gaps after filtering
- Integration test: filter state round-trips through URL query parameters across re-renders

## Effort

- 0.75 weeks

## Priority

- HIGH

## Phase

- 32

# REQ-TUI-004: Dependency Graph Visualization

## Metadata
- **Category**: TUI
- **Subcategory**: View
- **Priority**: HIGH
- **Phase**: 26
- **Status**: MISSING
- **Dependencies**: REQ-TUI-001, REQ-API-004
- **Blocks**: (none)

## Requirement

The TUI shall provide a dependency graph view that renders the requirement
dependency DAG using box-drawing characters, with color-coded status
indicators and keyboard navigation between nodes.

## Rationale

Dependency visualization is essential for understanding blocking chains,
critical paths, and parallelizable work. An ASCII graph renderer provides
this directly in the terminal without requiring a browser or external tool.
This complements the `rtmx deps` text output with an interactive,
navigable visualization.

## Design

### Layout

The graph view uses a layered (Sugiyama-style) layout algorithm that
assigns each node to a horizontal layer based on its topological depth,
then minimizes edge crossings within each layer.

```
Layer 0          Layer 1          Layer 2          Layer 3
+----------+     +----------+     +----------+     +----------+
|REQ-GO-001|---->|REQ-GO-009|---->|REQ-GO-019|---->|REQ-GO-027|
|[COMPLETE]|     |[COMPLETE]|     |[COMPLETE]|     |[COMPLETE]|
+----------+     +----------+     +----------+     +----------+
                      |           +----------+
                      +---------->|REQ-GO-032|
                                  |[COMPLETE]|
                                  +----------+
```

### Color Coding

- Green: COMPLETE
- Yellow: PARTIAL
- Red: MISSING
- Dim: NOT_STARTED
- Bold border: currently selected node
- Highlighted edge: path from selected to root

### Navigation

| Key | Action |
|-----|--------|
| h/l or Left/Right | Move between layers |
| j/k or Up/Down | Move between nodes in a layer |
| Enter | Open detail pane for selected node |
| c | Toggle critical path highlight |
| f | Filter graph to selected node's subgraph |
| a | Show all (reset filter) |
| +/- | Zoom: expand/collapse node content |

### Rendering

Uses Lipgloss for box rendering and styling. The graph is rendered into a
2D character grid, then displayed via a bubbles/viewport for scrolling.
Large graphs (>50 visible nodes) collapse to category-level summaries with
expand-on-demand.

### Scope Control

By default, shows the full graph filtered to incomplete requirements and
their immediate dependencies. The `f` key narrows to a subgraph rooted
at the selected node. The `a` key restores the full view.

## Acceptance Criteria

1. Graph view renders dependency DAG with box-drawing characters.
2. Nodes are color-coded by status.
3. Keyboard navigation moves between nodes across layers.
4. Enter on a node opens its detail pane.
5. Critical path highlighting shows the longest path in bold.
6. Subgraph filtering works from any selected node.
7. Large graphs (>50 nodes) default to collapsed category view.
8. Viewport scrolling handles graphs wider/taller than terminal.
9. Edge crossings are minimized within each layer.

## Files to Create/Modify

- `internal/tui/views/graph.go` -- Graph view model and renderer
- `internal/tui/views/graph_test.go` -- Layout and render tests
- `internal/tui/layout/sugiyama.go` -- Layered graph layout algorithm
- `internal/tui/layout/sugiyama_test.go` -- Layout algorithm tests

## Effort Estimate

1 week

## Test Strategy

- Layout algorithm: verify layer assignment matches topological depth
- Edge crossing minimization: verify reduction vs naive ordering
- Render output: golden file tests for small known graphs
- Navigation: verify cursor movement across layers
- Subgraph filter: verify correct node subset after filtering
- Large graph collapse: verify category grouping at threshold

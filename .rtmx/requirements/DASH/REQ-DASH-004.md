# REQ-DASH-004: Interactive Dependency Graph

## Metadata
- **Category**: DASH
- **Subcategory**: View
- **Priority**: HIGH
- **Phase**: 28
- **Status**: MISSING
- **Dependencies**: REQ-DASH-001, REQ-API-004
- **Blocks**: (none)

## Requirement

The web dashboard shall provide an interactive dependency graph visualization
using a force-directed or hierarchical layout, with clickable nodes, zoom/pan,
status-based color coding, and critical path highlighting.

## Rationale

The dependency graph is the most visually impactful view in a project
management tool. An interactive graph lets users explore blocking chains,
identify bottlenecks, and understand the project structure spatially. The
web platform enables richer interaction (zoom, pan, hover tooltips) than
the terminal can provide.

## Design

### Technology: D3.js Force-Directed Graph

D3.js (v7, ~80KB gzipped) provides the graph rendering engine. The
force-directed layout naturally separates clusters and reveals structure.
An alternative hierarchical (dagre) layout is available via toggle for
users who prefer layered views.

### Implementation

```html
<!-- Graph container -->
<div id="graph" x-data="graphView()" x-init="init()">
  <svg width="100%" height="600">
    <!-- D3 renders nodes and edges here -->
  </svg>
  <div class="controls">
    <button @click="toggleLayout()">Layout: Force / Hierarchy</button>
    <button @click="highlightCriticalPath()">Critical Path</button>
    <button @click="resetZoom()">Reset Zoom</button>
  </div>
</div>
```

### Node Appearance

- Circle radius proportional to effort_weeks
- Fill color by status: green (COMPLETE), yellow (PARTIAL), red (MISSING), gray (NOT_STARTED)
- Border thickness indicates priority (P0 = thick, LOW = thin)
- Label: req_id (or truncated text at high zoom)

### Interactions

- **Click node**: Select, show tooltip with full details
- **Double-click node**: Navigate to requirement detail page
- **Hover edge**: Highlight the dependency chain
- **Zoom/Pan**: Mouse wheel + drag (D3 zoom behavior)
- **Filter**: Category and status dropdowns filter visible nodes
- **Critical path**: Toggle button highlights the longest path in bold red

### Data Source

Fetches from `GET /api/graph` (REQ-API-004). Category and status filters
use the endpoint's query parameters to reduce data before rendering.

### Performance

For graphs >200 nodes, the force simulation runs with a reduced tick count
and node collision radius to maintain 60fps. Nodes outside the viewport
are culled from rendering (not from simulation).

## Acceptance Criteria

1. Graph renders all requirements as nodes with edges for dependencies.
2. Nodes are color-coded by status and sized by effort.
3. Click on a node shows a detail tooltip.
4. Double-click navigates to the requirement detail page.
5. Zoom and pan work smoothly.
6. Category and status filters reduce visible nodes.
7. Critical path toggle highlights the longest dependency chain.
8. Force-directed and hierarchical layouts are both available.
9. Graph renders within 1 second for a 300-node graph.
10. Reset zoom button returns to default viewport.

## Files to Create/Modify

- `dashboard/graph.html` -- Graph view template
- `dashboard/js/graph.js` -- D3 graph renderer and interactions
- `dashboard/vendor/d3.min.js` -- Vendored D3.js v7

## Effort Estimate

1.5 weeks

## Test Strategy

- Data loading: verify correct node/edge count from API response
- Node coloring: verify status-to-color mapping
- Navigation: double-click fires correct route change
- Filter: verify node count after category filter applied
- Performance: render 300-node graph in < 1 second (benchmark)
- Critical path: verify highlighted path matches API metadata

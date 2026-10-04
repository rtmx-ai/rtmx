# REQ-DASH-080: Print-Optimized CSS for Graph View

## Summary

The graph view shall include print-optimized CSS via `@media print` rules that produce a clean,
readable printout of the dependency graph. Print styles use a white background, ensure all node
labels are visible, hide interactive controls and navigation, and scale the graph to fit the
page. A print button in the graph toolbar triggers `window.print()`.

## Acceptance Criteria

1. `styles.css` contains `@media print` rules that apply only when printing the graph view
2. Print styles set background to white and all text to black for maximum contrast
3. Print styles hide: navigation sidebar, toolbar buttons, zoom controls, toast container, command palette, and any overlay elements
4. Print styles ensure all node labels are visible (no text truncation, no opacity reduction)
5. The graph SVG scales to fit the print page width while maintaining aspect ratio (`max-width: 100%; height: auto`)
6. The graph title and legend are visible in print output
7. Edge paths print as solid black lines (no color-dependent meaning lost)
8. Node fill colors print correctly; a black-and-white fallback uses pattern fills (hatching for blocked, dots for in-progress) via `@media print` overrides
9. The graph toolbar contains a print button (printer icon) that calls `window.print()`
10. Page margins are set to 0.5in for clean printing on standard paper sizes

## Dependencies

- REQ-DASH-048 (dagre layout provides the graph SVG to print)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/styles.css` (@media print rules, pattern fill definitions)
- `internal/dashboard/templates/partials/graph.html` (print button in toolbar)

## Test Strategy

- Unit test: @media print rules exist in styles.css and target correct selectors
- Unit test: navigation, toolbar, and overlay elements have `display: none` in print context
- Unit test: SVG has `max-width: 100%` and `height: auto` in print context
- Unit test: print button calls window.print()
- Integration test: puppeteer/playwright print-to-PDF produces readable output with visible labels and legend
- Visual test: printed PDF compared against golden reference for a known graph

## Effort

- 0.50 weeks

## Priority

- LOW

## Phase

- 32

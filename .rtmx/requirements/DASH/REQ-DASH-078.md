# REQ-DASH-078: Export Graph as Standalone SVG File

## Summary

The graph toolbar shall include an "Export SVG" button that downloads the current graph view
as a standalone SVG file. The exported SVG includes all visual encoding (node colors, shapes,
edge paths, labels, legend, title) with inline styles rather than CSS class references,
ensuring the file renders correctly when opened in any SVG viewer or embedded in documents.

## Acceptance Criteria

1. Graph toolbar contains an "Export SVG" button with a download icon
2. Clicking the button downloads a file named `{project}-graph-{date}.svg` (e.g., `rtmx-graph-2026-06-09.svg`)
3. The exported SVG is a complete standalone document with `xmlns` and `viewBox` attributes
4. All CSS class-based styles are converted to inline `style` attributes on each element (no `<style>` block or external CSS references)
5. Node fill colors, stroke colors, font sizes, and edge path styles are all inlined
6. The legend (status color key) is included in the exported SVG, positioned in the bottom-right corner
7. A title element is included at the top of the SVG: "{project} Dependency Graph -- {date}"
8. The exported SVG reflects the current view state: applied filters, expanded/collapsed clusters, group-by mode
9. The exported SVG does not include interactive elements (hover handlers, click handlers, cursor styles)
10. Export of a 300-node graph completes in under 2 seconds

## Dependencies

- REQ-DASH-048 (dagre layout provides the rendered SVG to export)
- REQ-DASH-057 (legend must exist to be included in export)

## Blocks

- REQ-DASH-079 (PNG export uses the SVG export as its source)

## Files to Modify

- `internal/dashboard/static/app.js` (SVG export function: clone SVG, inline styles, trigger download)
- `internal/dashboard/templates/partials/graph.html` (export SVG button in toolbar)

## Test Strategy

- Unit test: export function produces valid SVG with xmlns and viewBox
- Unit test: exported SVG has no `<style>` block and no class attributes; all styling is inline
- Unit test: exported SVG includes legend group and title text
- Unit test: exported SVG excludes cursor, onclick, and onmouseover attributes
- Unit test: filename follows the naming convention with project name and date
- Integration test: download triggered in browser produces a file that parses as valid XML
- Performance test: 300-node graph export completes in under 2 seconds
- Visual test: exported SVG opened in browser matches the dashboard graph appearance

## Effort

- 1.00 weeks

## Priority

- MEDIUM

## Phase

- 32

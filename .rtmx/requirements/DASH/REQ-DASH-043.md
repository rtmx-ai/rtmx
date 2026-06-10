# REQ-DASH-043: Vendor dagre.js as Static Layout Engine

## Summary

Vendor dagre.js (the Sugiyama-based directed graph layout library) as a static JavaScript file in `internal/dashboard/static/vendor/`, alongside the existing d3.v7.min.js. dagre.js computes layered node positions and edge routing for DAGs, which the existing D3 renderer will consume for positioning instead of force simulation. No npm or Node.js build step is introduced; the file is committed as a static vendor asset and embedded via Go's embed.FS.

## Acceptance Criteria

1. `internal/dashboard/static/vendor/dagre.min.js` exists and contains a UMD or IIFE build of dagre (version >= 0.8.5).
2. The vendored file exposes a global `dagre` object when loaded in a browser without module bundlers.
3. `internal/dashboard/templates/layout.html` includes a `<script>` tag loading `dagre.min.js` before `app.js`.
4. `dagre.min.js` file size is under 300KB (uncompressed; dagre@0.8.5 bundles graphlib at ~280KB).
5. The embed.FS in `internal/dashboard/embed.go` includes the new vendor file without changes (wildcard glob already covers `static/vendor/`).
6. `dagre.graphlib.Graph` is constructable from browser console when dashboard is loaded, confirming the library is available.
7. No npm, no package.json, no node_modules directory is created.
8. Existing graph rendering continues to function unchanged (no regressions).

## Dependencies

- REQ-DASH-001 (SPA framework provides the embed.FS and vendor directory structure)

## Blocks

- REQ-DASH-048 (dagre layout replacement requires the library to be available)

## Files to Modify

- `internal/dashboard/static/vendor/dagre.min.js` (new file, vendored)
- `internal/dashboard/templates/layout.html` (add script tag)

## Test Strategy

- Unit test: verify `dagre.min.js` exists in the embedded filesystem via Go test on embed.FS
- Integration test: serve dashboard, confirm `dagre` global is defined via script evaluation
- Regression test: existing graph partial renders without errors after adding the new script

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 32

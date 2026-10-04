# REQ-DASH-093: Treemap Partial Template Skeleton

## Summary

Create the HTML partial template for the treemap view at `internal/dashboard/templates/partials/treemap.html`. The template provides the SVG container and a `<script type="application/json">` block where the Go handler injects hierarchical treemap data. Register the `/partials/treemap` route in the dashboard server. This is the structural skeleton only; D3 rendering logic is handled by REQ-DASH-087.

## Acceptance Criteria

1. The file `internal/dashboard/templates/partials/treemap.html` exists and is a valid Go HTML template.
2. The template contains an SVG container element with an id of `treemap-svg` and responsive sizing (16:9 aspect ratio, min-height 400px).
3. The template contains a `<script type="application/json" id="treemap-data">` block for server-injected JSON.
4. The `/partials/treemap` route is registered in serve_dashboard.go and returns the rendered partial.
5. The partial integrates with the view switcher (REQ-DASH-066) via htmx swap.
6. The template uses the same layout wrapper and design tokens as other partials.

## Dependencies

- REQ-DASH-066 (view switcher provides the partial swap mechanism)
- REQ-DASH-010 (design tokens for consistent styling)

## Blocks

- REQ-DASH-071 (Go handler needs the template to inject JSON into)
- REQ-DASH-087 (D3 rendering needs the SVG container)

## Files to Modify

- `internal/dashboard/templates/partials/treemap.html` (new file)
- `internal/cmd/serve_dashboard.go` (register /partials/treemap route)

## Test Strategy

- Unit test: GET /partials/treemap returns 200 with HTML containing the SVG container
- Unit test: response contains the treemap-data script block
- Integration test: view switcher loads treemap partial via htmx swap

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 32

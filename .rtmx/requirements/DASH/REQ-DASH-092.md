# REQ-DASH-092: Timeline Partial Template Skeleton

## Summary

Create the HTML partial template for the timeline view at `internal/dashboard/templates/partials/timeline.html`. The template provides the SVG container, zoom controls, and a `<script type="application/json">` block where the Go handler injects timeline data. Register the `/partials/timeline` route in the dashboard server. This is the structural skeleton only; D3 rendering logic is handled by REQ-DASH-086.

## Acceptance Criteria

1. The file `internal/dashboard/templates/partials/timeline.html` exists and is a valid Go HTML template.
2. The template contains an SVG container element with an id of `timeline-svg` and responsive sizing.
3. The template contains zoom control buttons (zoom in, zoom out, reset) consistent with graph view controls.
4. The template contains a `<script type="application/json" id="timeline-data">` block for server-injected JSON.
5. The `/partials/timeline` route is registered in serve_dashboard.go and returns the rendered partial.
6. The partial integrates with the view switcher (REQ-DASH-066) via htmx swap.
7. The template uses the same layout wrapper and design tokens as other partials.

## Dependencies

- REQ-DASH-066 (view switcher provides the partial swap mechanism)
- REQ-DASH-010 (design tokens for consistent styling)

## Blocks

- REQ-DASH-067 (Go handler needs the template to inject JSON into)
- REQ-DASH-086 (D3 rendering needs the SVG container)

## Files to Modify

- `internal/dashboard/templates/partials/timeline.html` (new file)
- `internal/cmd/serve_dashboard.go` (register /partials/timeline route)

## Test Strategy

- Unit test: GET /partials/timeline returns 200 with HTML containing the SVG container
- Unit test: response contains the timeline-data script block
- Integration test: view switcher loads timeline partial via htmx swap

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 32

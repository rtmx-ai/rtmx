# REQ-DASH-027: Req-ID Links Across All Views Open Panel via htmx

## Summary

All clickable requirement ID elements across all dashboard views shall open the detail
panel via htmx, loading the requirement detail into the panel body.

## Acceptance Criteria

1. All `.req-id` clickable elements use `hx-target="#detail-panel-body"` to load detail content
2. Clicking a req-id sets `panelOpen=true` via Alpine `@click`
3. Works in requirements table, kanban cards, releases table, and health checks views
4. Panel body content is fetched from the detail partial endpoint

## Dependencies

- REQ-DASH-026 (panel slide animation)

## Blocks

- REQ-DASH-028

## Files to Modify

- `internal/dashboard/templates/partials/requirements.html`
- `internal/dashboard/templates/partials/kanban.html`
- `internal/dashboard/templates/partials/releases.html`
- `internal/dashboard/templates/partials/detail.html`

## Test Strategy

- Unit test: req-id elements have correct hx-target and @click attributes
- Unit test: clicking req-id in requirements table opens panel with correct content
- Integration test: panel opens from kanban, releases, and health views

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 31

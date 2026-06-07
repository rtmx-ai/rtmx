# REQ-DASH-012: Dark Theme Component Styling

## Summary

All dashboard components (tables, badges, kanban, progress bars, forms) shall be
restyled to match the rtmx.ai dark theme using design tokens.

## Acceptance Criteria

1. Data tables: header bg `#0f0f0f`, row borders `var(--rtmx-border)`, hover `var(--rtmx-bg-elevated)`
2. Status badges: pill shape, uppercase, semibold, color-on-transparent-bg pattern:
   - Complete: bg `rgba(16, 185, 129, 0.2)`, color `var(--rtmx-emerald)`
   - Partial: bg `rgba(245, 158, 11, 0.2)`, color `#f59e0b`
   - Missing: bg `rgba(239, 68, 68, 0.2)`, color `#ef4444`
   - Not Started: bg `rgba(107, 114, 128, 0.2)`, color `var(--rtmx-text-muted)`
3. Priority indicators use left-border pattern with updated colors
4. Kanban columns: bg `var(--rtmx-bg)`, cards bg `var(--rtmx-bg-card)`, border `var(--rtmx-border)`
5. Progress bars: track `var(--rtmx-border)`, fill gradient `#10b981 -> #6ee7b7`
6. Buttons: primary bg `var(--rtmx-emerald)`, text `var(--rtmx-bg)`, hover `var(--rtmx-emerald-light)`
7. CTA button hover glow: `box-shadow: 0 4px 20px rgba(16, 185, 129, 0.4)`
8. Form inputs: bg `var(--rtmx-bg-card)`, border `var(--rtmx-border)`, focus border `var(--rtmx-emerald)`
9. Graph legend uses updated status colors matching badge colors

## Dependencies

- REQ-DASH-010 (design tokens)
- REQ-DASH-011 (dark layout)

## Files to Modify

- `internal/dashboard/static/styles.css`
- `internal/dashboard/templates/partials/status.html`
- `internal/dashboard/templates/partials/requirements.html`
- `internal/dashboard/templates/partials/kanban.html`
- `internal/dashboard/templates/partials/graph.html`
- `internal/dashboard/templates/partials/releases.html`
- `internal/dashboard/templates/partials/health.html`

## Test Strategy

- Unit test: styles.css contains no hardcoded light-theme hex values (#f9fafb as bg, white cards)
- Unit test: all badge classes reference token-based colors
- Visual: screenshot comparison of each partial against rtmx.ai reference
- Regression: all existing dashboard tests pass

## Effort

- 1.5 weeks

## Priority

- HIGH

## Phase

- 30

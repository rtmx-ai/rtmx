# REQ-DASH-010: Design System Tokens from rtmx.ai

## Summary

The `rtmx serve` dashboard shall use CSS custom properties (design tokens) extracted
from the rtmx.ai website, replacing all hardcoded color values with semantic variables.

## Acceptance Criteria

1. `styles.css` defines the following CSS custom properties on `:root`:
   - `--rtmx-emerald: #10b981` (primary brand color)
   - `--rtmx-emerald-light: #6ee7b7` (primary hover/highlight)
   - `--rtmx-bg: #0a0a0a` (page background)
   - `--rtmx-bg-card: #171717` (card/surface background)
   - `--rtmx-bg-elevated: #1f1f1f` (elevated surface)
   - `--rtmx-border: #262626` (default borders)
   - `--rtmx-border-hover: #404040` (border hover state)
   - `--rtmx-text: #f9fafb` (primary text)
   - `--rtmx-text-muted: #9ca3af` (secondary text)
   - `--rtmx-text-dim: #6b7280` (tertiary/disabled text)
   - Status colors: complete (#22c55e), partial (#f59e0b), missing (#ef4444), not-started (#6b7280)
2. All color references in `styles.css` use `var(--rtmx-*)` instead of hardcoded hex values
3. All color references in HTML templates use CSS classes backed by design tokens, not inline hex
4. The dark theme matches the rtmx.ai website visual identity

## Dependencies

- REQ-DASH-001 (SPA framework)

## Blocks

- REQ-DASH-011, REQ-DASH-012, REQ-DASH-013

## Files to Modify

- `internal/dashboard/static/styles.css`
- `internal/dashboard/templates/layout.html` (remove Tailwind CDN, use token-based classes)
- `internal/dashboard/templates/partials/*.html` (replace inline color references)

## Test Strategy

- Unit test: parse `styles.css` and verify all `--rtmx-*` custom properties are defined
- Visual test: screenshot comparison of dashboard pages against rtmx.ai reference
- Regression: existing embed tests continue to pass

## Effort

- 1.0 week

## Priority

- HIGH

## Phase

- 30

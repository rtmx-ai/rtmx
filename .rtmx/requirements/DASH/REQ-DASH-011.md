# REQ-DASH-011: Dark Theme Layout and Navigation

## Summary

The dashboard layout and navigation shall use the rtmx.ai dark theme: dark background,
emerald accent navigation, and consistent card/surface hierarchy.

## Acceptance Criteria

1. Page background uses `var(--rtmx-bg)` (#0a0a0a)
2. Navigation bar uses `var(--rtmx-bg-card)` (#171717) background
3. Active nav link uses emerald accent (`var(--rtmx-emerald)`)
4. Cards use `var(--rtmx-bg-card)` background with `var(--rtmx-border)` border
5. Card hover state changes border to `var(--rtmx-emerald)` or `var(--rtmx-border-hover)`
6. Primary text uses `var(--rtmx-text)` (#f9fafb)
7. Secondary/label text uses `var(--rtmx-text-muted)` (#9ca3af)
8. Font stack includes 'JetBrains Mono' for monospace elements (req IDs, code)
9. System font stack for body text matches rtmx.ai
10. Focus indicators use `outline: 2px solid var(--rtmx-emerald)` per accessibility

## Dependencies

- REQ-DASH-010 (design tokens)

## Blocks

- REQ-DASH-012, REQ-DASH-013

## Files to Modify

- `internal/dashboard/templates/layout.html`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: layout HTML contains dark-theme CSS classes
- Unit test: no `bg-gray-50`, `text-gray-900`, or light-theme Tailwind classes remain
- Regression: all existing SPA embed tests pass

## Effort

- 1.0 week

## Priority

- HIGH

## Phase

- 30

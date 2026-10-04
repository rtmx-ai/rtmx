# REQ-DASH-014: Toast Container and Alpine.js Store

## Summary

The layout shell shall contain a toast notification container powered by an Alpine.js store,
providing a reusable showToast(type, message) function for success, error, and info feedback
across all dashboard views.

## Acceptance Criteria

1. `layout.html` contains a `#toast-container` element wired to Alpine `$store.toasts`
2. `app.js` defines `Alpine.store('toasts', ...)` with a `showToast(type, message)` method
3. Toast types render with distinct colors: success (emerald), error (red), info (blue)
4. Toasts auto-dismiss after 4 seconds
5. Maximum 5 toasts visible simultaneously; oldest removed when limit exceeded
6. Toasts animate in with a slide-in transition

## Dependencies

- REQ-DASH-001 (SPA framework)
- REQ-DASH-010 (design tokens)

## Blocks

- REQ-DASH-015

## Files to Modify

- `internal/dashboard/templates/layout.html`
- `internal/dashboard/static/app.js`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: verify Alpine store exposes showToast function and manages toast array
- Integration test: trigger showToast and verify DOM element appears with correct class
- Regression: existing layout embed tests continue to pass

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 31

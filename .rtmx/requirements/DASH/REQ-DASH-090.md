# REQ-DASH-090: URL State Restore and pushState/popState Integration (082b)

## Summary

Deserialize URL query parameters on page load to restore graph settings, and integrate with browser history via popstate events. Loading a URL with graph query parameters reproduces the exact graph state. Browser back/forward buttons navigate through graph state history. The detail panel URL (REQ-DASH-029) takes precedence when the panel is open.

## Acceptance Criteria

1. Loading a URL with graph query parameters (e.g., `?groupBy=status&filter=BLOCKED`) initializes the graph with those settings.
2. Browser back button restores the previous graph state (view, groupBy, filters, zoom, etc.) via popstate event.
3. Browser forward button re-applies the state that was navigated away from.
4. Copying the URL and opening it in a new tab reproduces the exact graph state.
5. URL state integrates with the existing pushState pattern from REQ-DASH-029 (detail panel URL takes precedence when panel is open).
6. A URL with no query parameters loads the graph with all defaults.
7. Invalid parameter values in the URL are silently ignored; defaults are used.
8. Back/forward navigation cycles correctly through multiple graph state changes.

## Dependencies

- REQ-DASH-082 (URL encoding format must be defined for deserialization to consume)
- REQ-DASH-029 (existing pushState/popstate URL pattern for detail panel)

## Blocks

- None currently identified

## Files to Modify

- `internal/dashboard/static/app.js` (URL deserialization on load, popstate handler for graph state)
- `internal/dashboard/templates/partials/graph.html` (read URL params on partial load to initialize settings)

## Test Strategy

- Unit test: loading URL with `?groupBy=status&filter=BLOCKED` initializes graph with those settings
- Unit test: popstate event restores graph settings from the popped URL
- Unit test: URL with no query parameters loads graph with all defaults
- Unit test: invalid parameter values fall back to defaults without error
- Integration test: full round-trip -- set state, copy URL, load in new context, verify identical graph
- Integration test: back/forward navigation cycles through multiple graph state changes

## Effort

- 0.50 weeks

## Priority

- MEDIUM

## Phase

- 32

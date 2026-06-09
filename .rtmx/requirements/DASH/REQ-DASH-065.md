# REQ-DASH-065: Search-to-Highlight with Command Palette Integration

## Summary

A search box in the graph toolbar shall allow users to type a query that highlights
matching nodes by requirement ID or description text. Matching nodes receive an emerald
glow effect, and the viewport pans to center the first match. The search integrates with
the command palette (REQ-DASH-031): prefixing a palette query with `>graph ` scopes the
search to graph nodes and applies the same highlight-and-pan behavior.

## Acceptance Criteria

1. A search input field appears in the graph toolbar with placeholder text "Search graph..."
2. Typing in the search box filters nodes by substring match against requirement ID and description (case-insensitive)
3. Matching nodes receive an emerald glow effect (`--color-emerald-500` box-shadow or stroke, 3px)
4. Non-matching nodes dim to 0.4 opacity when a search query is active
5. The viewport pans and zooms to center the first matching node with 200ms animated transition
6. Pressing Enter or down-arrow cycles through matches, panning to each in sequence
7. Pressing Escape or clearing the search input restores all nodes to full opacity
8. The command palette (Cmd+K) supports `>graph <query>` as a scoped command that applies the same search-to-highlight behavior
9. Match count is displayed next to the search input (e.g., "3 of 12 matches")
10. Search state clears when the graph data changes (filter, group-by, or data reload)

## Dependencies

- REQ-DASH-048 (dagre layout engine renders searchable node elements with ID and description data)
- REQ-DASH-031 (command palette modal provides the `>` prefix command routing infrastructure)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/graph.js`
- `internal/dashboard/static/app.js`
- `internal/dashboard/templates/partials/graph.html`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: typing a query highlights nodes matching by ID substring
- Unit test: typing a query highlights nodes matching by description substring (case-insensitive)
- Unit test: first match is centered in the viewport with animated pan
- Unit test: Enter key cycles through matches sequentially
- Unit test: Escape clears search and restores all nodes to full opacity
- Unit test: command palette `>graph REQ-DASH` triggers graph search with correct query
- Integration test: search-to-highlight with real graph data and pan animation

## Effort

- 0.75 weeks

## Priority

- MEDIUM

## Phase

- 32

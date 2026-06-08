# REQ-DASH-033: Arrow Key Navigation in Palette Results with Enter to Select

## Summary

The command palette result list shall support arrow key navigation with visual highlight,
and Enter shall open the detail panel for the highlighted result.

## Acceptance Criteria

1. Up/Down arrow keys move highlight through the result list
2. Enter opens the detail panel for the highlighted result and closes the palette
3. Highlight wraps around from last to first and vice versa
4. Currently highlighted item is visually distinct (elevated background)

## Dependencies

- REQ-DASH-032 (palette search and results)
- REQ-DASH-028 (panel close)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/app.js`

## Test Strategy

- Unit test: Down arrow moves highlight to next result
- Unit test: Up arrow moves highlight to previous result
- Unit test: Enter on highlighted result opens detail panel
- Unit test: highlight wraps at list boundaries

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 31

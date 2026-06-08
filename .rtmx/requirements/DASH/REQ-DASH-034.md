# REQ-DASH-034: Palette Action Prefixes for Navigation and Filtering

## Summary

The command palette shall support action prefixes (>, @, #) to switch between
navigation commands, assignee filtering, and category filtering modes.

## Acceptance Criteria

1. `>` prefix shows page navigation commands (go to Status, Requirements, Graph, Kanban, Releases, Health, Agents)
2. `@` prefix shows unique assignees as filter options
3. `#` prefix shows categories as filter options
4. Selecting a navigation command navigates to that page
5. Selecting a filter option applies it to the requirements view

## Dependencies

- REQ-DASH-032 (palette search and results)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/app.js`

## Test Strategy

- Unit test: `>` prefix shows navigation commands
- Unit test: `@` prefix shows assignee list
- Unit test: `#` prefix shows category list
- Unit test: selecting navigation command changes page
- Unit test: selecting filter option applies filter to requirements view

## Effort

- 0.25 weeks

## Priority

- MEDIUM

## Phase

- 31

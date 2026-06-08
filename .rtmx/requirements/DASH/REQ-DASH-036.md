# REQ-DASH-036: Global g-Prefix Shortcuts for Page Navigation

## Summary

The dashboard shall support g-prefix keyboard shortcuts for quick page navigation,
where pressing g followed by a second key within 500ms navigates to the corresponding page.

## Acceptance Criteria

1. g then s navigates to Status page
2. g then r navigates to Requirements page
3. g then d navigates to Graph page
4. g then k navigates to Kanban page
5. g then v navigates to Releases page
6. g then h navigates to Health page
7. g then a navigates to Agents page
8. 500ms timeout for the second key; if timeout expires, the sequence resets

## Dependencies

- REQ-DASH-001 (SPA framework)

## Blocks

- REQ-DASH-037, REQ-DASH-038

## Files to Modify

- `internal/dashboard/static/app.js`

## Test Strategy

- Unit test: g then s navigates to status page
- Unit test: g then r navigates to requirements page
- Unit test: second key after 500ms timeout does not navigate
- Unit test: non-mapped second key resets sequence

## Effort

- 0.25 weeks

## Priority

- MEDIUM

## Phase

- 31

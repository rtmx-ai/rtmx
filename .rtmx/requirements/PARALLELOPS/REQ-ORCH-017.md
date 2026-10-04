# REQ-ORCH-017: Dashboard Parallel Workstream Visualization

## Metadata
- **Category**: ORCH
- **Subcategory**: ParallelOps
- **Priority**: MEDIUM
- **Phase**: 32
- **Status**: MISSING
- **Dependencies**: REQ-ORCH-012, REQ-ORCH-016
- **Blocks**: (none)

## Summary

RTMX shall provide a dashboard visualization of parallel workstreams as a swimlane diagram. Each row represents one parallel group (execution phase). Within each row, webs are displayed as blocks showing requirement count, total effort, and a progress bar (complete/total requirements). Dependency arrows connect blocks across rows. Webs with file surface overlaps show dashed borders as a conflict warning. Active worktree assignments are indicated with the agent ID. The visualization uses htmx for dynamic updates and D3.js for the dependency arrows, consistent with the existing dashboard architecture.

## Acceptance Criteria

1. A new dashboard partial `parallel.html` renders the swimlane visualization.
2. Each parallel group is a horizontal row labeled "Phase N".
3. Each web block shows: web ID, requirement count, effort (weeks), and a progress bar.
4. Dependency arrows connect web blocks from lower phases to higher phases.
5. Webs with file surface overlaps have a dashed border and tooltip listing shared files.
6. Active worktree assignments show the agent ID badge on the web block.
7. The visualization updates via htmx polling (every 5 seconds).
8. The API endpoint `/api/parallel-plan` returns the structured data for the visualization.
9. The visualization is accessible from the dashboard navigation.

## Dependencies
- REQ-ORCH-012 (parallel groups provide the data model for swimlanes)
- REQ-ORCH-016 (worktree state provides active agent assignments)

## Blocks
- (none)

## Files to Modify
- internal/dashboard/templates/partials/parallel.html (new partial template)
- internal/dashboard/static/js/parallel.js (D3.js dependency arrows, optional)
- internal/cmd/serve_dashboard.go (add /api/parallel-plan endpoint and partial handler)
- internal/cmd/serve_integration_test.go (test endpoint response and HTML structure)

## Test Strategy
- Integration test verifying `/api/parallel-plan` returns valid JSON with expected structure.
- Integration test verifying the parallel partial HTML contains expected elements (phase rows, web blocks).
- Manual visual testing with a multi-web database fixture.

## Effort
- 1.5 weeks

## Priority
- MEDIUM

# REQ-ORCH-013: CLI rtmx webs Command

## Metadata
- **Category**: ORCH
- **Subcategory**: ParallelOps
- **Priority**: HIGH
- **Phase**: 32
- **Status**: MISSING
- **Dependencies**: REQ-ORCH-010, REQ-ORCH-011, REQ-ORCH-012, REQ-ORCH-001, REQ-ORCH-007
- **Blocks**: (none)

## Summary

RTMX shall provide a new CLI command `rtmx webs` that displays all independent work webs with comprehensive metadata for planning parallel execution. The table output shows: web ID (0-indexed), requirement count, unblocked count, total effort (weeks), file surface (unique TestModule values), overlap warnings (which other webs share files), parallel group assignment, and cross-web dependency info. The `--json` flag outputs structured JSON for programmatic consumption by agents and scripts.

## Acceptance Criteria

1. `rtmx webs` displays a formatted table with columns: Web, Reqs, Unblocked, Effort, Files, Overlaps, Group, Depends On.
2. Webs with file surface overlaps show a warning indicator and list the overlapping web IDs.
3. The parallel group column shows which execution group each web belongs to.
4. The "Depends On" column lists upstream web IDs.
5. `rtmx webs --json` outputs a JSON array of web objects with all fields.
6. JSON output includes: id, requirements (list of IDs), unblocked (list of IDs), blocked (list of IDs), total_effort, file_surface (list of paths), overlaps (list of web IDs), parallel_group, depends_on (list of web IDs).
7. When no incomplete requirements exist, the command prints a message indicating all work is complete.
8. Exit code 0 on success.

## Dependencies
- REQ-ORCH-001 (DetectWebs provides the web list)
- REQ-ORCH-007 (DetectOverlaps provides file overlap data)
- REQ-ORCH-010 (WebDependencies provides cross-web dependency info)
- REQ-ORCH-011 (MergeOrder provides ordering context)
- REQ-ORCH-012 (ParallelGroups provides group assignments)

## Blocks
- (none)

## Files to Modify
- internal/cmd/webs.go (new command implementation)
- internal/cmd/webs_test.go (new test file)
- internal/cmd/root.go (register webs command)

## Test Strategy
- Unit tests using CommandContext with mock database covering: empty database, single web, multiple independent webs, webs with dependencies, webs with overlaps, JSON output format validation.
- Golden file tests for table output formatting.
- Integration test verifying command registration and help text.

## Effort
- 1.0 weeks

## Priority
- HIGH

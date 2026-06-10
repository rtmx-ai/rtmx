# REQ-ORCH-015: CLI rtmx merge-gate Command

## Metadata
- **Category**: ORCH
- **Subcategory**: ParallelOps
- **Priority**: HIGH
- **Phase**: 32
- **Status**: MISSING
- **Dependencies**: REQ-ORCH-010, REQ-ORCH-011, REQ-ORCH-006
- **Blocks**: (none)

## Summary

RTMX shall provide a new CLI command `rtmx merge-gate --web N` that validates whether a specific web is safe to merge. The command checks three conditions: (a) all requirements in the web have status COMPLETE, (b) all upstream webs (those this web depends on) have already been merged (no active claims or worktrees), and (c) no file surface conflicts exist with currently in-progress webs. Exit code 0 means safe to merge; exit code 1 means blocked, with a human-readable explanation of which conditions failed. The `--json` flag outputs structured JSON with pass/fail status and details.

## Acceptance Criteria

1. `rtmx merge-gate --web N` exits 0 when all three conditions are met.
2. The command exits 1 when any requirement in the web is not COMPLETE, listing the incomplete requirement IDs.
3. The command exits 1 when an upstream web has active claims, listing the upstream web ID and its active claims.
4. The command exits 1 when a web with file surface overlap is currently in-progress, listing the conflicting web ID and shared files.
5. When multiple conditions fail, all failures are reported (not just the first).
6. `rtmx merge-gate --web N --json` outputs JSON with: web_id, safe (bool), checks (array of {name, passed, details}).
7. The command returns an error if `--web` specifies a non-existent web index.
8. The command uses the claims system to determine in-progress status.

## Dependencies
- REQ-ORCH-010 (cross-web dependencies determine upstream webs)
- REQ-ORCH-011 (merge order provides context for sequencing)
- REQ-ORCH-006 (existing merge command validates web completion)

## Blocks
- (none)

## Files to Modify
- internal/cmd/merge_gate.go (new command implementation)
- internal/cmd/merge_gate_test.go (new test file)
- internal/cmd/root.go (register merge-gate command)

## Test Strategy
- Unit tests using CommandContext with mock database and claims covering: all checks pass, incomplete requirements, upstream web still active, file conflict with in-progress web, multiple failures at once, invalid web index, JSON output format.
- Integration test with actual claims.json file to verify claim-system integration.

## Effort
- 1.0 weeks

## Priority
- HIGH

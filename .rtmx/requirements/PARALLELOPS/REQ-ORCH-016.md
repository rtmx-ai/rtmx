# REQ-ORCH-016: Worktree Metadata Tracking

## Metadata
- **Category**: ORCH
- **Subcategory**: ParallelOps
- **Priority**: MEDIUM
- **Phase**: 32
- **Status**: MISSING
- **Dependencies**: REQ-ORCH-005, REQ-ORCH-008
- **Blocks**: REQ-ORCH-015, REQ-ORCH-017

## Summary

RTMX shall extend the orchestration claims system to track worktree assignments in a persistent metadata file `.rtmx/worktrees.json`. This file records which agent has which web in which worktree path, on which branch, and when the assignment was created. New functions `AssignWorktree`, `GetWorktreeState`, and `ListActiveWorktrees` provide the API. The worktree state is used by merge-gate to determine which webs are in-progress and by the dashboard to visualize active workstreams.

## Acceptance Criteria

1. `AssignWorktree(webID int, agentID string, worktreePath string, branch string) error` creates an entry in `.rtmx/worktrees.json`.
2. `GetWorktreeState(webID int) (*WorktreeAssignment, error)` returns the assignment for a specific web, or nil if not assigned.
3. `ListActiveWorktrees() ([]WorktreeAssignment, error)` returns all active worktree assignments.
4. `ReleaseWorktree(webID int) error` removes the worktree assignment.
5. The `WorktreeAssignment` struct includes: WebID, AgentID, WorktreePath, Branch, CreatedAt, LastHeartbeat.
6. File operations use atomic write (write to temp, rename) to prevent corruption.
7. Concurrent access is safe via file locking consistent with the existing claims system.
8. Worktree state survives process restarts (persisted to disk).

## Dependencies
- REQ-ORCH-005 (claims protocol provides the file locking pattern)
- REQ-ORCH-008 (worktree creation provides the git worktree management)

## Blocks
- REQ-ORCH-015 (merge-gate uses worktree state to check in-progress webs)
- REQ-ORCH-017 (dashboard visualizes active worktree assignments)

## Files to Modify
- internal/orchestration/worktree.go (new file with WorktreeAssignment type and functions)
- internal/orchestration/worktree_test.go (new test file)

## Test Strategy
- Unit tests for CRUD operations: assign, get, list, release.
- Concurrency test: multiple goroutines assigning/releasing worktrees simultaneously.
- Persistence test: assign, close manager, reopen, verify state preserved.
- Error cases: assign to already-assigned web, release non-existent, corrupted JSON file recovery.

## Effort
- 1.0 weeks

## Priority
- MEDIUM

# REQ-ORCH-008: Worktree Binding for Claimed Webs

## Metadata
- **Category**: ORCH
- **Subcategory**: Orchestration
- **Priority**: HIGH
- **Phase**: 20
- **Status**: MISSING
- **Dependencies**: REQ-ORCH-004
- **Blocks**: (none)

## Requirement
`rtmx next --batch --worktree` shall automatically create a git worktree for the claimed web, record the worktree path and branch in the claim metadata, and provide the agent an isolated workspace.

## Design
- Creates: `git worktree add .worktrees/web-{ID} -b agent/web-{ID}`
- Records worktree path and branch in claims.json
- `rtmx merge` (REQ-ORCH-006) cleans up the worktree after merge

## Acceptance Criteria
1. --worktree creates a git worktree at .worktrees/web-{ID}.
2. Branch name follows agent/web-{ID} convention.
3. Worktree path and branch recorded in claim metadata.
4. Worktree is a valid git checkout with full source.
5. Multiple concurrent worktrees do not conflict.

## Files to Create/Modify
- internal/cmd/next.go
- internal/cmd/next_test.go
- internal/orchestration/worktree.go
- internal/orchestration/worktree_test.go

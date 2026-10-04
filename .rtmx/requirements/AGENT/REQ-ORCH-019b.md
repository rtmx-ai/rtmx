# REQ-ORCH-019b: Falling edge is a requirement merge

## Metadata
- **Category**: ORCH
- **Subcategory**: AgentLoop
- **Priority**: HIGH
- **Phase**: 36
- **Status**: COMPLETE
- **Dependencies**:
- **Blocks**: REQ-ORCH-019|REQ-ORCH-019a
- **ADR**:

## Requirement

The loop SHALL advance when a requirement falls from in-progress to
merged, and SHALL NOT advance because a clock fired. A merge is either
a pull request merged whose title or body contains a requirement ID, or
a local status transition of that ID into COMPLETE since the stored
cursor. The watcher persists the cursor so a restart does not replay
old merges.

## Rationale

A 12-minute sleep fired whether or not anything landed, and it missed
the moment a requirement actually finished. The useful edge is the
falling edge: work just completed, so the next unblocked requirement
may now be free, and it should be decomposed before anyone codes.

## Acceptance Criteria

1. [ ] Cursor file lives at `.rtmx/loop/cursor.json` (gitignored via `.rtmx/.gitignore` entry `loop/`). It records the last seen merge SHA or PR number and the last seen COMPLETE set.
2. [ ] Given two fixture events where only the second is a newly merged PR naming `REQ-EX-002`, the watcher reports one edge for `REQ-EX-002` and none for the older PR.
3. [ ] A COMPLETE transition in the database that was not in the cursor is an edge, even without a PR (local `rtmx verify --update`).
4. [ ] A still-open PR is not an edge.
5. [ ] A timer-only wake with no new merge and no new COMPLETE returns zero edges.
6. [ ] PR discovery is behind an interface. Default implementation may shell out to `gh pr list --state merged`; tests inject a fake and do not call GitHub.
7. [ ] Tests bind `REQ-ORCH-019b`.

## Files

- `internal/cmd/loop_edge.go`
- `internal/cmd/loop_edge_test.go`
- `.rtmx/.gitignore` (document the `loop/` ignore in init as well; init change itself is REQ-ORCH-021)

## Out of Scope

- Webhooks inbound to a public server.
- Treating "PR opened" as an edge. Open is the opposite of the falling edge.

# REQ-ORCH-022: rtmx next skips requirements with an open PR

## Metadata
- **Category**: ORCH
- **Subcategory**: AgentLoop
- **Priority**: MEDIUM
- **Phase**: 36
- **Status**: COMPLETE
- **Dependencies**:
- **Blocks**: REQ-ORCH-019c
- **ADR**:

## Requirement

`rtmx next` SHALL skip an incomplete requirement when an open pull
request already names that requirement ID, and SHALL select the next
unblocked requirement that has no such PR. The skip is visible in the
output so an agent does not think the requirement vanished.

## Rationale

The delivery loop claimed requirements that already had open pull
requests, then stacked conflicting branches. The falling-edge tick
must not hand an agent work that is already in review.

## Acceptance Criteria

1. [ ] PR lookup is an interface. Tests inject open PR titles; production may use `gh pr list --state open`.
2. [ ] A requirement ID is "named" when it appears as a substring of the PR title or body.
3. [ ] Given REQ-A (open PR) blocking nothing and REQ-B unblocked with no PR, `rtmx next` selects REQ-B and prints that REQ-A was skipped because of an open PR.
4. [ ] If every unblocked requirement has an open PR, next reports idle / none, not the skipped ID as the selection.
5. [ ] When the lookup interface is unavailable (no `gh`, error), next keeps today's behavior and warns once. It does not fail the command.
6. [ ] `--json` includes `skipped` with `{req_id, reason: "open_pr"}`.
7. [ ] Tests bind `REQ-ORCH-022`.

## Files

- `internal/cmd/next.go`
- `internal/cmd/next_test.go`

## Out of Scope

- Closing or merging the open PR.
- Matching PRs by branch name only. Title or body ID match is the contract.

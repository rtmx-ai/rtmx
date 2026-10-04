# REQ-ORCH-023a: Warn-first delivery check

## Metadata
- **Category**: ORCH
- **Subcategory**: AgentLoop
- **Priority**: P0
- **Phase**: 37
- **Status**: COMPLETE
- **Dependencies**: REQ-ORCH-023|REQ-ORCH-021
- **Blocks**: REQ-ORCH-023b|REQ-ORCH-023c
- **ADR**:

## Requirement

`rtmx delivery-check` SHALL report whether the current branch maps to
exactly one requirement and whether each commit since the merge-base
names that requirement and an acceptance-criterion marker. Default mode
warns (exit 0). `--strict` fails (exit 1) on violations.

## Acceptance Criteria

1. [ ] `rtmx delivery-check [--req ID] [--base REF] [--strict] [--json]` exists.
2. [ ] Without `--req`, the PR title/body (or `git log` subject aggregation when no PR metadata) must contain exactly one `REQ-…` ID; with `--req`, that ID is forced.
3. [ ] Each commit since merge-base must contain the requirement ID and an AC marker (`AC1`, `AC-1`, or `ac_id=…`).
4. [ ] Default: warnings printed, exit 0. `--strict`: exit 1 when any check fails.
5. [ ] `--json` emits `{ok, warnings[], errors[], req_id, commits[]}`.
6. [ ] Tests use fixture git repos (no network) and bind `REQ-ORCH-023a`.

## Files

- `internal/cmd/delivery_check.go`
- `internal/cmd/delivery_check_test.go`

## Out of Scope

- Hard-reject pre-commit hooks as default install.

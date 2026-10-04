# REQ-ORCH-020: rtmx decompose splits a coarse requirement

## Metadata
- **Category**: ORCH
- **Subcategory**: AgentLoop
- **Priority**: HIGH
- **Phase**: 36
- **Status**: COMPLETE
- **Dependencies**:
- **Blocks**: REQ-ORCH-019c
- **ADR**:

## Requirement

`rtmx decompose REQ-ID` SHALL turn a coarse requirement into child
requirements when it is not already atomic, and SHALL no-op when it is.
Children are real CSV rows and requirement files. The parent depends on
the children (parent stays incomplete until they are COMPLETE). Each
child is sized so an agent can deliver it as one pull request, with one
commit per acceptance criterion.

## Rationale

Decomposition in the delivery session was an agent judgment call with
no command, so the next session cannot repeat it. The loop's falling
edge needs a deterministic pass: either the requirement is already a
single PR, or it becomes children that are.

## Atomic rule

A requirement is already atomic when all of the following hold:

- It has at least one acceptance criterion.
- It has no more than 5 acceptance criteria.
- Its requirement text does not say it is a parent that closes only when children complete.
- It does not already list children in `blocks` that are incomplete.

Otherwise it is coarse.

## Acceptance Criteria

1. [ ] `rtmx decompose REQ-ID` on an atomic fixture prints `already atomic` and does not add CSV rows.
2. [ ] On a coarse fixture (more than 5 ACs, or an explicit `## Decomposition` section listing child titles), it writes child requirement files and CSV rows with status MISSING, dependencies pointing at any named predecessors, and the parent `dependencies` updated to include the children.
3. [ ] Child IDs are `{parent}a`, `{parent}b`, … skipping IDs that already exist. A second run is idempotent: no duplicate children.
4. [ ] Each child file has testable acceptance criteria copied or split from the parent, not an empty checklist.
5. [ ] `--dry-run` prints the plan and writes nothing.
6. [ ] Refuses to decompose a COMPLETE requirement (exit non-zero).
7. [ ] `rtmx cycles` still reports no cycle after a decompose of the fixture.
8. [ ] Tests bind `REQ-ORCH-020`.

## Files

- `internal/cmd/decompose.go`
- `internal/cmd/decompose_test.go`

## Out of Scope

- LLM-authored splits. v1 uses the deterministic rules above. An agent may edit child text after the files exist.
- Implementing the children.

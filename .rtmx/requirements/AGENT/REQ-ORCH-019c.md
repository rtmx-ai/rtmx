# REQ-ORCH-019c: Each edge runs next, then decomposition

## Metadata
- **Category**: ORCH
- **Subcategory**: AgentLoop
- **Priority**: HIGH
- **Phase**: 36
- **Status**: COMPLETE
- **Dependencies**: REQ-ORCH-020|REQ-ORCH-022
- **Blocks**: REQ-ORCH-019|REQ-ORCH-019a
- **ADR**:

## Requirement

On each falling edge the loop SHALL run one tick: select the next
unblocked requirement (`rtmx next` / claim rules, including the open-PR
skip in REQ-ORCH-022), then run a decomposition pass on that
requirement (REQ-ORCH-020) before any implementation. The tick prints a
machine-readable plan the agent can follow: claimed ID, whether
decomposition wrote children, and the child IDs to implement one PR
each.

## Rationale

`rtmx next` alone hands an agent a possibly coarse requirement. The
failures in the delivery session were oversized requirements and
duplicate work on rows that already had an open PR. The edge must
drive decomposition, not another undifferentiated "go implement this."

## Acceptance Criteria

1. [ ] `rtmx loop --once` (with an injected edge) claims at most one requirement and invokes decompose on that ID.
2. [ ] If decompose splits the requirement, the plan lists child IDs and states that the parent is not the implementation target.
3. [ ] If decompose reports the requirement is already atomic, the plan says so and names that ID as the single PR target.
4. [ ] The tick does not create commits, branches, or pull requests.
5. [ ] JSON plan on stdout (or `--json`) includes `req_id`, `decomposed` (bool), `children` (array), `pr_policy` (`one_pr_per_requirement`), `commit_policy` (`one_commit_per_ac`).
6. [ ] When next returns nothing, the plan is `{"idle":true}` and the claim store is unchanged.
7. [ ] Tests bind `REQ-ORCH-019c`.

## Files

- `internal/cmd/loop_tick.go`
- `internal/cmd/loop_tick_test.go`

## Out of Scope

- Writing the child requirement bodies (that is REQ-ORCH-020).
- Deciding merge vs wait. The agent rule (REQ-ORCH-021) covers delivery shape; this tick only plans.

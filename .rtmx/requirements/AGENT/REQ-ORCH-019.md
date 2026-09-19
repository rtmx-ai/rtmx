# REQ-ORCH-019: Deployable falling-edge agent delivery loop

## Metadata
- **Category**: ORCH
- **Subcategory**: AgentLoop
- **Priority**: HIGH
- **Phase**: 36
- **Status**: COMPLETE
- **Dependencies**: REQ-ORCH-019a|REQ-ORCH-019b|REQ-ORCH-019c
- **Blocks**:
- **ADR**:

## Requirement

A user SHALL be able to deploy an RTMX agent loop that keeps working the
backlog. The loop advances on the falling edge of a requirement merge
(the requirement just became done), not on a timer. Each advance runs
`rtmx next` and then a decomposition pass on the requirement that next
selects, before any implementation commits.

## Rationale

The session loop that drained the backlog was a shell `sleep` plus an
agent prompt. That is not something a user can install. A clock also
fires when nothing merged, and it does not force decomposition. The
product needs a deployable unit whose trigger is "a requirement landed."

## Acceptance Criteria

1. [ ] Parent closes only when 019a (deployable runner), 019b (merge edge), and 019c (next then decompose) are COMPLETE.
2. [ ] Documented deploy path a user can run without editing this monorepo's agent session (command and/or unit file from `rtmx loop`).
3. [ ] A tick does not start because a timer elapsed. It starts because 019b observed a new merge, or because this is the bootstrap tick on an empty cursor with a non-empty backlog.
4. [ ] The tick's first mutations are `rtmx next` (claim) and decomposition (019c / REQ-ORCH-020). Implementation commits are out of band, after that pass.
5. [ ] When `rtmx next` reports no unblocked incomplete requirement, the loop waits. It does not busy-spin and it does not invent work.
6. [ ] Does not require managed-sync OAuth or Stripe. Local RTM is sufficient.

## Files

- `internal/cmd/loop.go` (019a)
- `internal/cmd/loop_edge.go` (019b)
- `internal/cmd/loop_tick.go` (019c)
- `features/agent_delivery_loop.feature`

## Out of Scope

- Hosting the coding agent (Cursor, Claude, or otherwise). The loop emits the next claim and the decomposition; the agent executes.
- Merging pull requests automatically.
- OAuth, Stripe, or website deploy.

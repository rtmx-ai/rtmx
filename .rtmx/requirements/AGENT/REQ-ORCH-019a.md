# REQ-ORCH-019a: rtmx loop is a deployable runner

## Metadata
- **Category**: ORCH
- **Subcategory**: AgentLoop
- **Priority**: HIGH
- **Phase**: 36
- **Status**: COMPLETE
- **Dependencies**: REQ-ORCH-019b|REQ-ORCH-019c
- **Blocks**: REQ-ORCH-019
- **ADR**:

## Requirement

`rtmx loop` SHALL run the delivery loop in the foreground, and
`rtmx loop install` SHALL write a host unit the user can enable so the
loop survives a terminal close. The runner calls the falling-edge
watcher (019b) and, on each edge, runs one tick (019c).

## Rationale

Users need a command they can deploy, not a prompt buried in one
coding session. Install must be explicit so `rtmx init` does not start
a background agent by surprise.

## Acceptance Criteria

1. [ ] `rtmx loop` blocks, watches for merge edges, and runs exactly one tick per edge.
2. [ ] `rtmx loop install` writes a unit file and prints the enable command. It does not enable the unit itself.
   - macOS: a LaunchAgent plist under the project or `~/Library/LaunchAgents`, documented.
   - Linux: a systemd user unit, documented.
   - The unit's `Exec` is `rtmx loop` with the project directory.
3. [ ] `rtmx loop install --dry-run` prints the unit and writes nothing.
4. [ ] Second `rtmx loop install` without `--force` refuses to overwrite an existing unit.
5. [ ] On bootstrap (no stored cursor) with at least one unblocked incomplete requirement, the runner performs one tick, then waits for edges.
6. [ ] When the backlog is empty, `rtmx loop` waits (or exits 0 with `--once` and a clear "idle" line). It does not exit non-zero.
7. [ ] `--once` performs at most one tick and exits 0, for tests and CI.
8. [ ] Tests bind `REQ-ORCH-019a`. No network, no real launchctl/systemctl.

## Files

- `internal/cmd/loop.go`
- `internal/cmd/loop_test.go`
- `internal/cmd/loop_install.go`

## Out of Scope

- Windows service installer.
- Starting Cursor or Claude as a child process.
- GitHub-hosted runners. A user may point the same `rtmx loop` at a self-hosted runner; that wiring is documentation, not a new service.

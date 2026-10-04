# REQ-ORCH-021: rtmx init installs the delivery rule

## Metadata
- **Category**: ORCH
- **Subcategory**: AgentLoop
- **Priority**: HIGH
- **Phase**: 36
- **Status**: COMPLETE
- **Dependencies**:
- **Blocks**:
- **ADR**:

## Requirement

`rtmx init` SHALL install an agent rule that enforces one commit per
acceptance criterion and one pull request per requirement. The rule is
the canonical text below. Init writes it where the detected coding
agents will read it, without starting a loop and without clobbering
unrelated instructions.

## Rationale

The existing agent prompt says "one requirement per cycle" and "commit
with the REQ-ID." It does not stop an agent from stuffing every
acceptance criterion into one commit or several requirements into one
pull request. That rule has to be present at project creation, not
pasted in later by whoever remembers.

## Canonical rule text

```
RTMX delivery rule:

- One pull request per requirement. The PR title contains the requirement ID.
- One commit per acceptance criterion of that requirement. The commit message names the requirement ID and the AC.
- Do not implement a parent in the same PR as its children. If `rtmx decompose` split the requirement, implement the children.
- Do not open a second PR for a requirement that already has an open PR.
- Mark the requirement COMPLETE only via `rtmx verify`, never by editing status by hand.
- Merge the PR only after verify is green for that requirement.
- Run `rtmx delivery-check` (warn-first) before opening or merging a PR; use `--strict` in release gates when the project opts in.
- When a requirement is ambiguous or multi-option, open a trade under `.rtmx/trades/` and resolve it before coding; prefer MCP `loop_tick` / `trade_*` / `delivery_check` over shelling arbitrary CLI.
```

## Acceptance Criteria

1. [ ] `rtmx init` writes `.rtmx/agent/delivery.md` containing the canonical rule, including both "One pull request per requirement" and "One commit per acceptance criterion."
2. [ ] If a `CLAUDE.md` is created or already exists in the project, init inserts the rule in a marked block `<!-- rtmx:delivery-rule -->` … `<!-- /rtmx:delivery-rule -->`. A second init with `--force` replaces that block and leaves text outside the block unchanged.
3. [ ] Init writes `.cursor/rules/rtmx-delivery.mdc` with the same rule body so Cursor loads it as a project rule.
4. [ ] Init does not enable `rtmx loop` and does not write a LaunchAgent or systemd unit (that is `rtmx loop install`).
5. [ ] `--dry-run` prints the paths and writes nothing. Init of an existing `.rtmx/` without `--force` still refuses, matching current init behavior; the rule files are part of the structure init creates, so a fresh init always includes them.
6. [ ] Tests bind `REQ-ORCH-021`.

## Files

- `internal/cmd/init.go`
- `internal/cmd/init_test.go`
- embedded template for `.rtmx/agent/delivery.md` and `.cursor/rules/rtmx-delivery.mdc`

## Out of Scope

- Installing rules into every agent in `rtmx install --agents` (Claude hooks, Copilot, Cline). Those stay on `rtmx install`. Init covers the project-local Cursor rule, Claude marked block, and the canonical file.
- Hard-reject git hooks as default. Mechanical warn-first check is REQ-ORCH-023a (`rtmx delivery-check`).

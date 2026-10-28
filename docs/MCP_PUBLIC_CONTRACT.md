# MCP public contract

**Status:** public surface for marketplace plugins and agent emitters  
**Traces:** REQ-MONO-062a · [REQ-MCP-006](../.rtmx/requirements/MCP/REQ-MCP-006.md) (stdio)  
**Slice:** monorepo `system/features/agent-marketplace/`

This document freezes the **agent-facing MCP payload** that Claude, Cursor,
Codex, and other clients install via marketplace plugins or `rtmx plugin emit`.
It does not change tool implementations — names, launch command, and semver
rules only.

## Launch (marketplace payload)

Plugins and MCP configs MUST spawn:

```bash
rtmx mcp-server --stdio
```

| Field | Value |
|-------|--------|
| Transport | JSON-RPC 2.0 over stdin/stdout (MCP stdio) |
| Command | `rtmx` (on `PATH`) |
| Args | `mcp-server`, `--stdio` |
| Prerequisite | `rtmx` binary installed (Homebrew / scoop / release / `go install`) |

HTTP mode (`rtmx mcp-server` without `--stdio`) remains supported for local
dashboards; **marketplace packs use stdio only**.

Stdio behavior and parity with HTTP are specified by **REQ-MCP-006**. Schema
and discovery coverage live in `internal/adapters/mcp/server_test.go`
(`tools/list`, stdio RPC tests).

## Tool catalog

Canonical names as returned by MCP `tools/list`. Renaming or removing a tool
is a **breaking** change (see Semver).

### Core RTM tools

| Tool | Role |
|------|------|
| `status` | Completion counts / percentages |
| `backlog` | Prioritized incomplete requirements |
| `health` | Database health checks |
| `deps` | Dependency info (one req or overview) |
| `verify` | Run tests; map results to requirements |
| `markers` | Requirement markers in tests |
| `next` | Work webs / highest-priority unblocked work |
| `claim` | Claim a requirement (mutation; needs `agent_id`) |
| `release` | Release a claim (mutation; needs `agent_id`) |
| `release_assign` | Assign reqs to a version (mutation) |
| `set_status` | Set status with provenance (mutation; rejects COMPLETE) |

### Scientific workflow tools

| Tool | Role |
|------|------|
| `loop_tick` | Next + claim + decompose plan for one agent cycle |
| `decompose` | Split a coarse requirement into children |
| `hygiene` | RTM hygiene findings |
| `cycles` | Circular dependency report |
| `webs` | Independent work webs |
| `context` | Token-efficient RTM summary |
| `delivery_check` | Warn-first one-PR / one-commit delivery check |
| `trade_open` | Open a trade-analysis checkpoint |
| `trade_list` | List trade checkpoints |
| `trade_resolve` | Resolve a trade with a choice |

Input schemas are those advertised in `tools/list` (JSON Schema on each tool).
Additive optional properties are backward compatible; removing a required
property or changing a property’s type is breaking.

## Semver (CLI ↔ marketplace plugins)

MCP is part of the RTMX **public surface** (see `CLAUDE.md` Versioning Policy).

| Change | CLI bump | Plugin pack |
|--------|----------|-------------|
| Remove/rename tool; remove required input; incompatible schema/type change | **MAJOR** | Coordinated **MAJOR** (or explicit break note) |
| New tool; new optional input; new compatible output field | **MINOR** | Track CLI **minor** (bump pack minor) |
| Bugfix with no schema/name change | **PATCH** | Pack bump optional |

**Marketplace plugins track the CLI minor version.** A plugin built for
`v0.N` MUST work with CLI `v0.N.*` patch releases. Crossing a CLI MAJOR that
breaks MCP requires a new plugin major and marketplace refresh.

## Non-goals (this contract)

- Changing tool handler behavior or adding tools in the 062a docs-only child
- Vendoring the Go binary inside a plugin archive
- Replacing stdio with a remote HTTP MCP for local CSV databases

## Consumers

- Shared agent-plugin pack: [`plugins/rtmx/`](../plugins/rtmx/) (`mcp.json` is
  the only marketplace MCP command/args source of truth — REQ-MONO-062b)
- Claude / Cursor marketplace wrappers (must consume that pack)
- `rtmx plugin emit` / Codex `mcp_servers.rtmx` configs

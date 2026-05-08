# RTMX

**Track what you built, what's tested, and what's next -- from the terminal.**

[![CI](https://github.com/rtmx-ai/rtmx/actions/workflows/ci.yml/badge.svg)](https://github.com/rtmx-ai/rtmx/actions/workflows/ci.yml)
[![Coverage Status](https://coveralls.io/repos/github/rtmx-ai/rtmx/badge.svg?branch=main)](https://coveralls.io/github/rtmx-ai/rtmx?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/rtmx-ai/rtmx)](https://goreportcard.com/report/github.com/rtmx-ai/rtmx)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

<!-- TODO: Replace with docs/assets/rtmx-workflow.gif once generated from docs/tapes/workflow.tape -->
<!-- ![RTMX workflow](docs/assets/rtmx-workflow.gif) -->

RTMX is a CLI that manages requirements traceability as a CSV file in git.
Every requirement has an ID, a spec, and linked tests. Status is derived
from test results -- not manually updated. AI agents query your requirements
via MCP and build against your intent, not their own.

## Install

```bash
brew install rtmx-ai/tap/rtmx        # macOS / Linux
scoop install rtmx                    # Windows
go install github.com/rtmx-ai/rtmx/cmd/rtmx@latest
```

Or download a binary from [releases](https://github.com/rtmx-ai/rtmx/releases).

## What It Does

| Command | What it does |
|---------|-------------|
| `rtmx status` | Completion dashboard across all requirements and phases |
| `rtmx next --one` | Pick the highest-priority unblocked requirement |
| `rtmx verify` | Run tests and cross-reference against requirements |
| `rtmx health` | Lint your RTM: orphaned tests, circular deps, stale refs |
| `rtmx backlog` | Prioritized work items with critical path analysis |
| `rtmx mcp-server` | 7 tools for AI agents over JSON-RPC (read + write) |

33 commands total. Run `rtmx --help` for the full list.

## The AI Workflow

```mermaid
flowchart TD
    A["Requirements\n(CSV in git)"] --> B["rtmx next\nPick unblocked requirement"]
    B --> C["Agent writes\ncode + tests"]
    C --> D{"rtmx verify\nTests pass?"}
    D -- yes --> E["Status updates\nautomatically"]
    D -- no --> C
    E --> F["rtmx status\nTeam sees progress"]
    F -.-> B

    style A fill:#064e3b,stroke:#10b981,color:#fff
    style B fill:#064e3b,stroke:#10b981,color:#fff
    style C fill:#1c1917,stroke:#6ee7b7,color:#fff
    style D fill:#78350f,stroke:#fbbf24,color:#fff
    style E fill:#064e3b,stroke:#10b981,color:#fff
    style F fill:#064e3b,stroke:#10b981,color:#fff
```

An agent runs `rtmx next --one`, gets a specific requirement, writes code
and tests against the spec, and `rtmx verify` confirms the work is done.
No human triages, assigns, or updates a ticket.

## Why CSV in Git

```mermaid
block-beta
    columns 1
    block:header["database.csv  --  Pull Request #42"]
        columns 1
    end
    block:removed["- REQ-AUTH-003,auth,mfa,TOTP-based MFA,...,missing,,"]
        columns 1
    end
    block:added["+ REQ-AUTH-003,auth,mfa,TOTP-based MFA,...,complete,test_totp_flow,"]
        columns 1
    end
    block:context["  REQ-AUTH-004,auth,session,Session timeout,...,partial,,"]
        columns 1
    end

    style header fill:#1c1917,stroke:#404040,color:#9ca3af
    style removed fill:#450a0a,stroke:#dc2626,color:#fca5a5
    style added fill:#052e16,stroke:#16a34a,color:#86efac
    style context fill:#1c1917,stroke:#404040,color:#9ca3af
```

- **Human-readable diffs** in PRs -- one row changed, one requirement done
- **Works everywhere** -- offline, air-gapped, no database, no API
- **AI agents parse it** without an SDK -- it's a file
- **git blame** tells you when and why every requirement changed
- **No vendor lock-in** -- it's your data in your repo

## MCP Integration

```mermaid
flowchart LR
    subgraph agents["AI Agents"]
        A1["Claude Code"]
        A2["Cursor"]
        A3["Custom Agent"]
    end

    subgraph mcp["rtmx mcp-server"]
        T1["status"]
        T2["backlog"]
        T3["next"]
        T4["verify"]
        T5["health"]
        T6["markers"]
        T7["deps"]
    end

    subgraph repo["Git Repository"]
        DB[".rtmx/database.csv"]
        Tests["Test files"]
    end

    A1 -- "JSON-RPC" --> mcp
    A2 -- "JSON-RPC" --> mcp
    A3 -- "JSON-RPC" --> mcp
    mcp --> DB
    mcp --> Tests

    style agents fill:#1c1917,stroke:#6ee7b7,color:#fff
    style mcp fill:#064e3b,stroke:#10b981,color:#fff
    style repo fill:#1c1917,stroke:#404040,color:#fff
    style A1 fill:#1c1917,stroke:#6ee7b7,color:#fff
    style A2 fill:#1c1917,stroke:#6ee7b7,color:#fff
    style A3 fill:#1c1917,stroke:#6ee7b7,color:#fff
    style DB fill:#064e3b,stroke:#10b981,color:#fff
    style Tests fill:#064e3b,stroke:#10b981,color:#fff
```

7 read-only tools plus mutation tools with agent authorization and atomic
claim/release for multi-agent coordination. Run `rtmx mcp-server` to start.

## Test Verification

`rtmx verify` auto-detects your test framework. No configuration needed.

Go, Python/pytest, Rust/Cargo, Node.js/npm, Java/Gradle, Java/Maven,
Elixir/Mix, Swift, Dart, Ruby -- 10+ frameworks supported.

## Dogfooding

RTMX manages its own requirements. 219 requirements across 24 phases,
auto-verified in CI on every push. Run `rtmx status` in this repo to see it.

[Read the backstory](https://rtmx.ai/blog/show-hn-rtmx) -- how and why this tool was built.

## Technical Details

- Single static binary -- Go, `CGO_ENABLED=0`, zero runtime dependencies
- Linux, macOS, Windows -- amd64 and arm64
- 2 external dependencies (Cobra + YAML parser)
- GPG-signed releases with SBOM
- Apache 2.0

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, building,
and testing instructions.

### Migrating from Python CLI

> **Deprecation Notice:** The Python `rtmx` CLI (`pip install rtmx`) is deprecated
> and will reach end-of-life on 2026-09-25.

Install the Go CLI, verify with `rtmx status`, then `pip uninstall rtmx`.
Full migration guide: [docs/MIGRATION.md](docs/MIGRATION.md).

## Support

- Documentation: [rtmx.ai](https://rtmx.ai)
- Issues: [github.com/rtmx-ai/rtmx/issues](https://github.com/rtmx-ai/rtmx/issues)
- Email: dev@rtmx.ai

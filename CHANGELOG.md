# Changelog

All notable changes to RTMX are documented in this file.
Format follows [Keep a Changelog](https://keepachangelog.com/).

## [1.4.0] - 2026-06-04

### Added

- **`rtmx hygiene` command** — reports requirement actionability and traceability
  hygiene findings (effort bounds, generic owners, missing test mappings, missing
  external IDs, generic acceptance criteria, and dependency cycles), with JSON
  output, strict mode, a configurable effort range, and a `hygeine` typo alias.
  (REQ-HYGIENE-001)

### Changed

- Bump dependencies: `spf13/cobra` 1.8.0 → 1.10.2; CI actions `setup-go` 5.5 → 6.4,
  `github-script` 8 → 9, and `goreleaser-action` 6 → 7.

### Fixed

- `TestStaleTestReferences` now parses the RTM database with `encoding/csv`, so
  quoted fields containing commas (e.g. a `target_value` listing items) no longer
  shift column alignment and produce false stale-reference failures.

## [1.3.0] - 2026-06-04

### Added

- **Configurable completeness policy** for closed-loop verification
  (`rtmx.completeness`). Projects can now declare a multi-dimensional definition
  of "done": `policy: combinations` marks a requirement COMPLETE only when its
  passing tests cover at least `min_combinations` distinct tuples of the
  configured `dimensions` (a subset of `scope`, `technique`, `env`); otherwise it
  is PARTIAL. The default `policy: simple` preserves the historical
  single-passing-test rule, so existing projects are unaffected. This composes
  with the `phoenix`, `do178c`, and `iso26262` schemas, whose dimensional markers
  can now gate completion. Applies to the cross-language results path
  (`rtmx verify --results`); the native go-test path keeps the simple rule.
  `require_all_pass` (default true) still downgrades COMPLETE to PARTIAL on any
  failing test. (REQ-VERIFY-009)
- **Configurable dimension vocabularies** (`rtmx.completeness.vocabulary`).
  Projects can extend the accepted `scope`/`technique`/`env` marker values for
  results validation, so schema-specific values (e.g. the `phoenix` schema's
  `static_field`/`dynamic_field`, or custom `.rtmx/schema.yaml` values) validate
  instead of being rejected. Built-in vocabularies are preserved when unset.
  (REQ-VERIFY-010)

## [1.0.0] - 2026-05-10

v1.0.0 signals API stability for the RTMX CLI. The public surface -- commands,
flags, CSV format, config schema, exit codes, and --json output -- is now
covered by semantic versioning guarantees.

### Added

- **42 CLI commands** covering the full requirements traceability lifecycle:
  init, setup, status, backlog, deps, cycles, health, verify, release (gate,
  assign, scope), plan, forecast, adapt, next, merge, tui, serve, install,
  context, reconcile, plugin, migrate, and more
- **Multi-agent orchestration** with atomic claim/release, worktree binding,
  batch claiming, and merge validation for parallel AI-assisted development
- **MCP server** with 10 tools (status, backlog, health, deps, cycles, next,
  verify, add, update, transition) for AI agent integration
- **Schema plugin framework** with 4 built-in schemas (core, phoenix, do178c,
  iso26262) and YAML custom schema support via .rtmx/schema.yaml
- **16 language scanners** (Go, Python, Rust, TypeScript, Java, C#, Ruby,
  Swift, Kotlin, Scala, C/C++, Zig, Elixir, Haskell, Lua, Bash) for
  cross-language requirement marker detection
- **Release planning** with version assignment, scope visualization, gate
  verification, and version policy enforcement
- **Forecasting** with Monte Carlo simulation, velocity tracking, confidence
  intervals, and burndown/burnup projections
- **GitHub and Jira adapters** for bidirectional issue synchronization
- **Database schema migration** with automatic column detection and upgrade
- **Verify --audit** for detecting stale references and false positives
- **GPG-signed releases** with SBOM (SPDX), checksums, and verification
- **Cross-platform distribution**: Linux/macOS/Windows (amd64+arm64),
  Homebrew tap, Scoop bucket, .deb/.rpm packages, Docker images, npm wrapper
- **Homebrew-core formula** (Formula/rtmx.rb) for `brew install rtmx`
- **APT repository tooling** (scripts/apt-repo.sh) for `apt install rtmx`
- **Claude Code integration** with hooks, context injection, and skill pack
- **10 AI agent configurations** (Claude, Cursor, Copilot, Cline, Windsurf,
  Aider, Amazon Q, Gemini, Zed, Continue)
- **OpenZiti zero-trust networking** support for air-gapped deployments
- **Web dashboard** (rtmx serve) and terminal dashboard (rtmx tui)
- **BDD feature specs** with Gherkin acceptance criteria
- **Benchmark framework** with regression detection, tolerance bands, and
  CI integration

### Changed

- README rewritten as Show HN landing page with workflow diagrams and
  install-to-value flow
- Version policy now enforces semantic versioning based on category-driven
  increment rules

### Migration from Python CLI

The Go CLI achieves full feature parity with the Python CLI. Migration:

```bash
pip uninstall rtmx        # Remove Python CLI
brew install rtmx-ai/tap/rtmx  # Install Go CLI
rtmx migrate              # Convert any Python-era config
```

See docs/MIGRATION.md for details.

## [0.3.0] - 2026-05-02

### Added

- Release planning commands (release gate, assign, scope)
- Version filtering and assignment
- Gate verification with pre-tag hook

## [0.2.0] - 2026-04-15

### Added

- Multi-language scanner support (16 languages)
- GitHub and Jira adapter integration
- Verify --audit for stale reference detection

## [0.1.0] - 2026-02-17

### Added

- Initial Go CLI release, architectural transition from Python
- Core commands: init, status, backlog, deps, cycles, health, verify
- CSV database parser with full Python format compatibility
- Graph algorithms (Tarjan SCC, topological sort, critical path)
- GoReleaser with signed binaries and SBOM

[1.0.0]: https://github.com/rtmx-ai/rtmx/releases/tag/v1.0.0
[0.3.0]: https://github.com/rtmx-ai/rtmx/releases/tag/v0.3.0
[0.2.0]: https://github.com/rtmx-ai/rtmx/releases/tag/v0.2.0
[0.1.0]: https://github.com/rtmx-ai/rtmx/releases/tag/v0.1.0

# REQ-DIST-002: Standalone Binary CLI Distribution

## Status: MISSING
## Priority: HIGH
## Phase: 13

## Description
RTMX CLI shall be distributed as standalone binaries for Linux, Windows, and macOS, requiring no runtime dependencies (Python, Node.js, etc.) for end users. This enables universal adoption across development environments without toolchain prerequisites.

## Rationale
The current Python-based CLI requires users to have Python installed and manage virtual environments. For RTMX to achieve universal adoption:
- Enterprise environments may restrict Python installation
- CI/CD runners benefit from single binary deployment
- Cross-platform consistency requires hermetic builds
- AI agent orchestration systems need lightweight, dependency-free tools

## Technology Selection: Go

Go is selected over Rust for the standalone CLI port based on:

| Factor | Go | Rust |
|--------|-----|------|
| Build time | ~10s | ~2-5min |
| Binary size | ~10MB | ~5MB |
| Cross-compile | Built-in (GOOS/GOARCH) | Requires cross toolchains |
| CLI ecosystem | Cobra, Viper (mature) | Clap (excellent but smaller ecosystem) |
| Learning curve | Moderate | Steep |
| Error handling | Explicit, simple | Complex (Result<T, E>) |
| Concurrency | Goroutines (simple) | async/await (complex) |
| Existing ecosystem | REQ-LANG-003 creates Go module | No existing RTMX Rust code |

**Decision**: Go provides faster iteration for CLI development while maintaining excellent cross-platform support.

## Acceptance Criteria
- [ ] Go CLI binary `rtmx` compiles for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- [ ] Single binary with zero runtime dependencies
- [ ] All core commands implemented: `status`, `backlog`, `health`, `validate`, `graph`, `from-tests`
- [ ] CSV parser reads/writes `rtm_database.csv` with full schema support
- [ ] YAML parser loads `rtmx.yaml` configuration
- [ ] Graph algorithms (Tarjan SCC, topological sort, critical path) ported from Python
- [ ] Output formatting matches Python CLI (tables, colors, progress bars)
- [ ] GitHub releases include pre-built binaries for all platforms
- [ ] Installation via: `go install github.com/rtmx-ai/rtmx-go/cmd/rtmx@latest`
- [ ] Homebrew formula for macOS: `brew install rtmx-ai/tap/rtmx`
- [ ] Scoop manifest for Windows
- [ ] `.deb` and `.rpm` packages for Linux distributions
- [ ] Shell completion scripts for bash, zsh, fish, PowerShell
- [ ] Binary size under 15MB per platform
- [ ] Startup time under 100ms for `rtmx --version`
- [ ] Feature parity with Python CLI v1.0

## Architecture

```
github.com/rtmx-ai/rtmx-go/
├── cmd/
│   └── rtmx/
│       └── main.go           # Cobra CLI entry point
├── internal/
│   ├── config/
│   │   ├── config.go         # RTMXConfig from rtmx.yaml
│   │   └── defaults.go       # Default configuration values
│   ├── database/
│   │   ├── csv.go            # CSV parser (encoding/csv)
│   │   ├── requirement.go    # Requirement struct
│   │   └── validation.go     # Database validation
│   ├── graph/
│   │   ├── dependency.go     # Tarjan's SCC, topo sort
│   │   └── critical_path.go  # Critical path analysis
│   ├── output/
│   │   ├── table.go          # Table formatting (tablewriter)
│   │   ├── color.go          # ANSI colors (fatih/color)
│   │   └── progress.go       # Progress bars (cheggaaa/pb)
│   └── cmd/
│       ├── status.go         # rtmx status
│       ├── backlog.go        # rtmx backlog
│       ├── health.go         # rtmx health
│       ├── validate.go       # rtmx validate
│       ├── graph.go          # rtmx graph
│       └── from_tests.go     # rtmx from-tests
├── pkg/
│   └── rtmx/                 # Public API for Go integration
│       ├── database.go
│       └── requirement.go
├── go.mod
├── go.sum
├── Makefile
└── .goreleaser.yaml          # Cross-platform release automation
```

## Dependencies (Go Modules)

| Library | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/spf13/viper` | Configuration management |
| `gopkg.in/yaml.v3` | YAML parsing |
| `github.com/olekukonko/tablewriter` | Table output |
| `github.com/fatih/color` | ANSI colors |
| `github.com/cheggaaa/pb/v3` | Progress bars |

## Release Process

1. **GoReleaser** automates cross-compilation and packaging
2. **GitHub Actions** triggers on version tags (`v*`)
3. **Artifacts uploaded** to GitHub Releases
4. **Homebrew tap** auto-updated via GoReleaser
5. **Checksums** (SHA256) published for verification

## Migration Path

| Phase | Scope | Notes |
|-------|-------|-------|
| 1 | Read-only commands | `status`, `backlog`, `health` - no database writes |
| 2 | Validation commands | `validate`, `graph` - analysis only |
| 3 | Write commands | `from-tests`, `init` - database modifications |
| 4 | Sync commands | `sync` - requires rtmx-sync integration |
| 5 | Full parity | All Python CLI features ported |

## Test Strategy

- **Unit tests**: Go testing package with table-driven tests
- **Integration tests**: Golden file comparison with Python CLI output
- **Cross-platform CI**: GitHub Actions matrix (ubuntu, macos, windows)
- **Benchmark tests**: Startup time, large database performance

## Test Cases
1. `tests/test_go_cli.py::test_binary_exists_all_platforms` - Binaries compile for all targets
2. `tests/test_go_cli.py::test_csv_round_trip` - CSV read/write preserves data
3. `tests/test_go_cli.py::test_status_output_matches_python` - Output parity check
4. `tests/test_go_cli.py::test_graph_algorithms_match` - Graph algorithm correctness
5. `tests/test_go_cli.py::test_startup_time_under_100ms` - Performance requirement
6. `tests/test_go_cli.py::test_zero_dependencies` - Binary has no runtime deps

## Dependencies
- REQ-LANG-007: Language-agnostic marker annotation spec (for marker parsing)
- REQ-LANG-003: Go testing integration (companion module)

## Blocks
- REQ-LANG-001: JavaScript plugin (blocks rtmx npm, not Go CLI)
- REQ-LANG-002: Java plugin (uses Go CLI for marker extraction)
- REQ-LANG-004: Rust plugin (uses Go CLI for marker extraction)
- REQ-LANG-005: C# plugin (uses Go CLI for marker extraction)
- REQ-LANG-006: Ruby plugin (uses Go CLI for marker extraction)

## Effort
8.0 weeks

## References
- ADR-006: npm Distribution Trade Study (context for TypeScript vs Go decision)
- GoReleaser documentation: https://goreleaser.com/
- Cobra CLI framework: https://cobra.dev/

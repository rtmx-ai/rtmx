# RTMX Go CLI

Requirements Traceability Matrix toolkit - Go implementation.

[![CI](https://github.com/rtmx-ai/rtmx/actions/workflows/ci.yml/badge.svg)](https://github.com/rtmx-ai/rtmx/actions/workflows/ci.yml)
[![Coverage Status](https://coveralls.io/repos/github/rtmx-ai/rtmx/badge.svg?branch=main)](https://coveralls.io/github/rtmx-ai/rtmx?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/rtmx-ai/rtmx)](https://goreportcard.com/report/github.com/rtmx-ai/rtmx)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## Overview

RTMX is a CLI tool for managing requirements traceability in GenAI-driven development. This is the Go implementation, providing:

- Single static binary (no runtime dependencies)
- Cross-platform support (Linux, macOS, Windows)
- Fast startup and execution
- Full feature parity with the Python CLI

## Installation

### Homebrew (macOS/Linux)

```bash
brew install rtmx-ai/tap/rtmx
```

### Scoop (Windows)

```powershell
scoop bucket add rtmx https://github.com/rtmx-ai/scoop-bucket
scoop install rtmx
```

### Go Install

```bash
go install github.com/rtmx-ai/rtmx/cmd/rtmx@latest
```

### Download Binary

Download the appropriate binary from the [releases page](https://github.com/rtmx-ai/rtmx/releases).

## Usage

```bash
# Show help
rtmx --help

# Check version
rtmx version

# Show RTM status
rtmx status

# Show backlog
rtmx backlog

# Run health check
rtmx health
```

## Migrating from Python CLI

> **Deprecation Notice:** The Python `rtmx` CLI (`pip install rtmx`) is deprecated
> and will reach end-of-life on 2026-09-25. Please migrate to the Go CLI.

The Go CLI is a drop-in replacement for the Python CLI. All configuration files,
database formats, and command outputs are fully compatible.

### Migration steps

1. Install the Go CLI (see [Installation](#installation) above).
2. Verify the Go CLI works with your project:
   ```bash
   rtmx status
   rtmx health
   ```
3. Remove the Python CLI:
   ```bash
   pip uninstall rtmx
   ```
4. (Optional) Use the built-in migration command:
   ```bash
   rtmx migrate --to-go
   ```

### What stays the same

- `.rtmx/` directory structure and `database.csv` format
- `rtmx.yaml` / `.rtmx/config.yaml` configuration files
- All CLI commands and flags
- JSON output schema
- Exit codes

### What is new

- Single static binary with no runtime dependencies
- Cross-platform support (Linux, macOS, Windows) from a single build
- Faster startup and execution
- Native Go test integration (`rtmx from-go`)

## Development

### Prerequisites

- Go 1.22+
- golangci-lint (for linting)
- goreleaser (for releases)

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run linter
make lint
```

### Testing

```bash
# Run all tests
make test

# Run parity tests against Python CLI
make parity

# Show coverage
make coverage
```

## Architecture

```
rtmx-go/
├── cmd/rtmx/           # Main entry point
├── internal/
│   ├── cmd/            # CLI commands
│   ├── config/         # Configuration management
│   ├── database/       # CSV/database operations
│   ├── graph/          # Dependency graph algorithms
│   ├── output/         # Formatting and display
│   ├── adapters/       # GitHub, Jira, MCP integrations
│   └── sync/           # CRDT sync and remotes
├── pkg/rtmx/           # Public API for Go integration
└── testdata/           # Test fixtures
```

## License

Apache 2.0 - See [LICENSE](LICENSE) for details.

## Support

- Documentation: https://rtmx.ai/docs
- Issues: https://github.com/rtmx-ai/rtmx/issues
- Email: dev@rtmx.ai

# RTMX Go CLI

Requirements Traceability Matrix toolkit - Go implementation.

[![CI](https://github.com/rtmx-ai/rtmx-go/actions/workflows/ci.yml/badge.svg)](https://github.com/rtmx-ai/rtmx-go/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rtmx-ai/rtmx-go)](https://goreportcard.com/report/github.com/rtmx-ai/rtmx-go)
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
go install github.com/rtmx-ai/rtmx-go/cmd/rtmx@latest
```

### Download Binary

Download the appropriate binary from the [releases page](https://github.com/rtmx-ai/rtmx-go/releases).

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

If you're currently using the Python CLI (`pip install rtmx`), you can migrate:

```bash
# Using the Python CLI migration tool
rtmx migrate --to-go

# Or install manually (see Installation above)
```

The Go CLI maintains full compatibility with the Python CLI's configuration files and database format.

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
- Issues: https://github.com/rtmx-ai/rtmx-go/issues
- Email: dev@rtmx.ai

# CLAUDE.md

This file provides guidance to Claude Code when working with the RTMX Go CLI codebase.

## Overview

This is the Go implementation of the RTMX CLI, providing a single static binary for requirements traceability management. It is a port of the Python CLI (`rtmx-ai/rtmx`).

## Quick Commands

```bash
make build        # Build the binary
make test         # Run tests
make lint         # Run linter
make dev          # Build with race detector
make build-all    # Build for all platforms
make parity       # Run parity tests against Python CLI
```

## Project Structure

```
rtmx-go/
├── cmd/rtmx/           # Main entry point
├── internal/
│   ├── cmd/            # CLI commands (Cobra)
│   ├── config/         # Configuration management (Viper)
│   ├── database/       # CSV parsing, Requirement model
│   ├── graph/          # Tarjan's SCC, topological sort, critical path
│   ├── output/         # Tables, colors, progress bars
│   ├── adapters/       # GitHub, Jira, MCP integrations
│   └── sync/           # CRDT sync and remotes
├── pkg/rtmx/           # Public API for Go integration
└── testdata/           # Golden files and fixtures
```

## Key Design Decisions

1. **Cobra for CLI** - Standard Go CLI framework with subcommands
2. **Viper for config** - YAML/JSON/ENV config management
3. **No CGO** - Pure Go for static binary distribution
4. **Internal packages** - Implementation details not exported
5. **Golden file tests** - Ensure output parity with Python CLI

## Development Workflow

### Adding a New Command

1. Create `internal/cmd/<command>.go`
2. Add command to `internal/cmd/root.go` in `init()`
3. Create test file `internal/cmd/<command>_test.go`
4. Add golden files to `testdata/` if needed

### Testing Parity

Every command must produce identical output to the Python CLI:

```bash
# Run parity tests
make parity

# Manual comparison
python -m rtmx status > /tmp/py.out
./bin/rtmx status > /tmp/go.out
diff /tmp/py.out /tmp/go.out
```

### Build & Release

```bash
# Local snapshot build
make snapshot

# Check release config
make release-check

# Release (triggered by tag)
git tag v0.1.0
git push origin v0.1.0
```

## Version Management

Version, commit, and date are injected via ldflags:

```go
// internal/cmd/root.go
var (
    Version = "dev"
    Commit  = "none"
    Date    = "unknown"
)
```

Build with:
```bash
go build -ldflags "-X github.com/rtmx-ai/rtmx-go/internal/cmd.Version=v0.1.0 ..."
```

## Compatibility Requirements

- **Config files**: Must read rtmx.yaml/.rtmx/config.yaml identically to Python
- **CSV format**: Must read/write rtm_database.csv identically to Python
- **Output format**: Tables, colors, progress bars must match Python
- **Exit codes**: Must match Python exit codes
- **JSON output**: Must match Python JSON schema exactly

## Common Patterns

### Error Handling

```go
func runCommand(cmd *cobra.Command, args []string) error {
    config, err := loadConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    // ...
}
```

### Color Output

```go
if !noColor && isTerminal() {
    output = colorize(output, "green")
}
```

### Progress Display

```go
bar := output.NewProgressBar(total)
for _, item := range items {
    process(item)
    bar.Increment()
}
bar.Finish()
```

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration
- `gopkg.in/yaml.v3` - YAML parsing

Minimize dependencies to keep binary size small (<15MB target).

## Testability Requirements

### Coverage Targets

| Package | Target | Enforcement |
|---------|--------|-------------|
| internal/adapters | 100% | Coveralls + CI gate |
| internal/cmd | 100% | Coveralls + CI gate |
| internal/database | 90% | Coveralls |
| internal/graph | 90% | Coveralls |
| internal/config | 90% | Coveralls |
| internal/output | 90% | Coveralls |
| cmd/rtmx | 100% | E2E tests |
| **Overall** | >90% | CI fails below 80% |

### Architecture Patterns for Testability

**ALL adapters and commands MUST be fully testable:**

1. **HTTPClient Interface** - All HTTP operations via interface
   ```go
   type HTTPClient interface {
       Do(req *http.Request) (*http.Response, error)
   }

   func NewGitHubAdapter(cfg *config.GitHubConfig, opts ...AdapterOption) (*GitHubAdapter, error)

   // Test with mock
   adapter, _ := NewGitHubAdapter(cfg, WithHTTPClient(mockClient))
   ```

2. **FileSystem Interface** - All file I/O via interface
   ```go
   type FileSystem interface {
       ReadFile(path string) ([]byte, error)
       WriteFile(path string, data []byte, perm os.FileMode) error
       Stat(path string) (os.FileInfo, error)
       MkdirAll(path string, perm os.FileMode) error
       Remove(path string) error
   }
   ```

3. **CommandContext** - Dependency injection for commands
   ```go
   type CommandContext struct {
       Config      *config.Config
       Database    *database.Database
       Output      io.Writer
       ErrOutput   io.Writer
       Input       io.Reader
       GetEnv      func(string) string
       FileSystem  FileSystem
   }
   ```

4. **Environment Abstraction** - Never call `os.Getenv` directly
   ```go
   // WRONG
   token := os.Getenv("GITHUB_TOKEN")

   // RIGHT
   func NewAdapter(cfg *Config, opts ...Option) *Adapter
   func WithEnvGetter(fn func(string) string) Option
   ```

### Test Patterns

1. **Table-Driven Tests** - Standard Go pattern for all test cases
   ```go
   func TestFeature(t *testing.T) {
       tests := []struct {
           name     string
           input    Input
           expected Output
           wantErr  bool
       }{...}

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {...})
       }
   }
   ```

2. **Golden File Tests** - For output formatting
   ```go
   golden.Assert(t, "status_output", actualOutput)
   // Update: go test -update
   ```

3. **Mock HTTP Server** - For adapter testing
   ```go
   server := testutil.NewMockServer()
   server.ExpectRequest("GET", "/repos/owner/repo/issues", MockResponse{...})
   defer server.Close()
   ```

4. **Fuzz Tests** - For parsing functions
   ```go
   func FuzzCSVParse(f *testing.F) {
       f.Add([]byte("header\nrow"))
       f.Fuzz(func(t *testing.T, data []byte) {
           ParseCSV(bytes.NewReader(data)) // must not panic
       })
   }
   ```

5. **Property-Based Tests** - For algorithm invariants
   ```go
   // Graph algorithms must satisfy invariants
   // - Tarjan finds ALL SCCs
   // - Topological sort respects ALL edges
   // - Critical path is actually critical
   ```

### Test Infrastructure Packages

```
internal/
├── adapters/
│   └── testutil/
│       ├── mock_server.go    # HTTP mock server
│       └── fixtures.go       # Adapter test data
├── testutil/
│   ├── fixtures.go           # Test database factories
│   ├── golden.go             # Golden file assertions
│   └── tempdir.go            # Temp directory helpers
└── cmd/
    └── testutil/
        └── context.go        # CommandContext factory
```

### CI Integration

```yaml
# Coverage is checked on every PR
- name: Run tests with coverage
  run: go test -coverprofile=coverage.out -covermode=atomic ./...

- name: Upload to Coveralls
  uses: coverallsapp/github-action@v2

- name: Check coverage threshold
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
    if (( $(echo "$COVERAGE < 80" | bc -l) )); then
      echo "Coverage $COVERAGE% is below 80% threshold"
      exit 1
    fi
```

### What NOT to Do

- **NO global state** - No package-level variables that affect behavior
- **NO direct os.Getenv** - Use injected environment getter
- **NO direct http.DefaultClient** - Use injected HTTPClient
- **NO direct file operations** - Use FileSystem interface
- **NO untested code paths** - Every error path must have a test

## Contact

- RTMX Engineering: dev@rtmx.ai
- Issues: https://github.com/rtmx-ai/rtmx-go/issues

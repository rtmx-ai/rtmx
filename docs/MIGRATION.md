# Migrating from Python CLI

> **Deprecation Notice:** The Python `rtmx` CLI (`pip install rtmx`) is deprecated
> and will reach end-of-life on 2026-09-25. Please migrate to the Go CLI.

The Go CLI is a drop-in replacement for the Python CLI. All configuration files,
database formats, and command outputs are fully compatible.

## Migration steps

1. Install the Go CLI:
   ```bash
   brew install rtmx-ai/tap/rtmx        # macOS / Linux
   scoop install rtmx                    # Windows
   go install github.com/rtmx-ai/rtmx/cmd/rtmx@latest
   ```

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

## What stays the same

- `.rtmx/` directory structure and `database.csv` format
- `rtmx.yaml` / `.rtmx/config.yaml` configuration files
- All CLI commands and flags
- JSON output schema
- Exit codes

## What is new

- Single static binary with no runtime dependencies
- Cross-platform support (Linux, macOS, Windows) from a single build
- Faster startup and execution
- Native Go test integration (`rtmx from-go`)

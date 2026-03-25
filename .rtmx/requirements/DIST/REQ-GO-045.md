# REQ-GO-045: Python CLI Deprecation Notice

## Metadata
- **Category**: DIST
- **Subcategory**: Deprecation
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Effort**: 1 week
- **Dependencies**: REQ-GO-073 (v0.1.0 release)
- **Blocks**: REQ-GO-046 (Migrate command)

## Requirement

The Go CLI distribution shall include a deprecation notice mechanism so that Python CLI users are informed of the migration path to the Go binary. This includes documentation, install script messaging, and a machine-readable deprecation manifest.

## Rationale

Users currently running `pip install rtmx` need a clear signal that the Python CLI is being superseded by the Go binary. The Go repo must provide the artifacts that enable this transition: a deprecation manifest, updated documentation, and install script that references the Go binary.

## Design

### Deprecation Manifest

Create `deprecation.json` in the repo root:

```json
{
  "deprecated": "rtmx (Python)",
  "successor": "rtmx (Go)",
  "install": "curl -fsSL https://rtmx.ai/install.sh | sh",
  "homebrew": "brew install rtmx-ai/tap/rtmx",
  "scoop": "scoop install rtmx",
  "migration_guide": "https://github.com/rtmx-ai/rtmx-go#migrating-from-python",
  "sunset_date": "2026-06-01"
}
```

### README Migration Section

Add a "Migrating from Python" section to README.md with:
- Install instructions for the Go binary
- Command mapping (any differences)
- Config file compatibility notes
- How to keep pytest plugin while using Go CLI

### Install Script Notice

The install script (`scripts/install.sh`) already exists and installs the Go binary. It should print a migration notice if it detects a Python rtmx installation.

## Acceptance Criteria

1. `deprecation.json` exists in repo root with successor info and install instructions
2. README.md contains a "Migrating from Python" section
3. Install script detects existing Python rtmx and prints migration notice
4. Deprecation manifest includes sunset date and all install methods

## Test Strategy

- **Test Module**: `test/deprecation_test.go`
- **Test Function**: `TestDeprecationNotice`
- **Validation Method**: System Test

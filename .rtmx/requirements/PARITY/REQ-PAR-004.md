# REQ-PAR-004: Health Command Parity

## Metadata
- **Category**: PARITY
- **Subcategory**: Health
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-012
- **Blocks**: REQ-GO-047

## Requirement

`rtmx health` shall match Python CLI health check behavior including missing checks, exit codes, and flags.

## Gaps to Close

### Missing Health Checks
1. `config_valid` - Validate config file syntax and required fields
2. `schema_valid` - Validate database schema (column names, types)
3. `agent_configs` - Check AI agent config files exist (CLAUDE.md, .cursorrules, etc.)

### Exit Code Fix
- Current Go: Warnings exit 1
- Python: Warnings exit 0, errors exit 2
- Fix: Warnings should exit 0 (or 1 with `--strict`)

### Missing Flags
- `--strict` - Treat warnings as errors (exit 1)
- `--check NAME` - Run only specific check(s)

## Acceptance Criteria

1. `config_valid` check validates YAML config
2. `schema_valid` check validates CSV column headers
3. `agent_configs` check detects CLAUDE.md, .cursorrules, copilot-instructions.md
4. Warnings exit 0 by default, exit 1 with `--strict`
5. `--check reciprocity` runs only the reciprocity check
6. Exit codes match Python behavior

## Files to Modify

- `internal/cmd/health.go` - Add checks, fix exit codes, add flags
- `internal/cmd/health_test.go` - Tests for new checks and flags

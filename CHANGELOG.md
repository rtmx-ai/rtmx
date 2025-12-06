# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.2] - 2025-12-06

### Added
- Unified `rtmx setup` command combining init, install, and makefile generation
- `rtmx health` command for project health scoring
- `rtmx diff` command for comparing RTM versions
- Phoenix-style backlog display with summary header and sections
- View modes for backlog: `--view all|critical|quick-wins|blockers`
- Colored terminal screenshots in README documentation
- Auto-regeneration of terminal screenshots in CI

### Changed
- Backlog output now matches Phoenix/Cyclone format with Critical Path and Quick Wins sections
- Phase display format changed to "Phase X" style
- Blocks column shows transitive count with direct count in parentheses
- Status icons updated (△ for PARTIAL)

### Improved
- Code coverage increased from 26% to 87%
- Added comprehensive test suite for CLI commands, models, adapters

### Fixed
- Type errors in health.py, setup.py, integrate.py
- Pre-commit hook compatibility issues

## [0.0.1] - 2024-12-03

### Added
- Initial release of RTMX
- Core RTM data model and CSV parser
- Dependency graph analysis with cycle detection
- Validation framework for requirement integrity
- CLI commands:
  - `rtmx init` - Initialize RTM structure
  - `rtmx status` - Show completion status with verbosity levels
  - `rtmx backlog` - Prioritized incomplete requirements
  - `rtmx deps` - Dependency visualization
  - `rtmx cycles` - Circular dependency detection
  - `rtmx reconcile` - Dependency consistency fixes
  - `rtmx from-tests` - Extract markers from test files
  - `rtmx makefile` - Generate Makefile targets
  - `rtmx analyze` - Project artifact discovery
  - `rtmx bootstrap` - Generate RTM from existing artifacts
  - `rtmx install` - Inject prompts into AI agent configs
  - `rtmx sync` - Bidirectional sync with GitHub/Jira
  - `rtmx mcp-server` - MCP protocol server
- Pytest plugin with custom markers:
  - `@pytest.mark.req()` - Link tests to requirements
  - Scope: `scope_unit`, `scope_integration`, `scope_system`
  - Technique: `technique_nominal`, `technique_parametric`, `technique_stress`
  - Environment: `env_simulation`, `env_hil`, `env_field`
- YAML configuration (rtmx.yaml)
- Service adapters for GitHub Issues and Jira
- MCP server for AI agent integration
- Agent prompt templates for Claude, Cursor, Copilot
- GitHub Actions CI/CD pipeline
- Pre-commit hooks (ruff, mypy)
- Security scanning (pip-audit, CodeQL)
- E2E test suite for lifecycle management

[Unreleased]: https://github.com/iotactical/rtmx/compare/v0.0.2...HEAD
[0.0.2]: https://github.com/iotactical/rtmx/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/iotactical/rtmx/releases/tag/v0.0.1

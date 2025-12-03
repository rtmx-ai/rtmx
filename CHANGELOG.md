# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- GitHub Actions CI/CD pipeline
- Pre-commit hooks configuration
- Security scanning with pip-audit
- SBOM generation (CycloneDX format)
- Dependabot for dependency updates
- CONTRIBUTING.md guide
- SECURITY.md policy
- Issue and PR templates

## [0.1.0] - 2024-12-03

### Added
- Initial release of RTMX
- Core RTM data model and CSV parser
- Dependency graph analysis with cycle detection
- Validation framework for requirement integrity
- CLI commands:
  - `rtmx status` - Show completion status with verbosity levels
  - `rtmx backlog` - Prioritized incomplete requirements
  - `rtmx deps` - Dependency visualization
  - `rtmx cycles` - Circular dependency detection
  - `rtmx reconcile` - Dependency consistency fixes
  - `rtmx from-tests` - Extract markers from test files
  - `rtmx init` - Initialize RTM structure
  - `rtmx makefile` - Generate Makefile targets
  - `rtmx analyze` - Project discovery
  - `rtmx bootstrap` - Initial RTM generation
  - `rtmx install` - Agent prompt injection
  - `rtmx sync` - Bidirectional sync with GitHub/Jira
  - `rtmx mcp-server` - MCP protocol server
- pytest plugin with custom markers:
  - `@pytest.mark.req()` - Link tests to requirements
  - Scope markers: `scope_unit`, `scope_integration`, `scope_system`
  - Technique markers: `technique_nominal`, `technique_stress`, etc.
- YAML configuration support (rtmx.yaml)
- Service adapters for GitHub Issues and Jira
- MCP (Model Context Protocol) server for AI agent integration
- Agent prompt templates for Claude Code, Cursor, and GitHub Copilot

[Unreleased]: https://github.com/iotactical/rtm/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/iotactical/rtm/releases/tag/v0.1.0

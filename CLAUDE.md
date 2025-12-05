# CLAUDE.md

This file provides guidance to Claude Code when working with the RTMX codebase.

## Quick Commands

```bash
make dev          # Install with dev dependencies
make test         # Run tests
make lint         # Run linter
make rtm          # Show RTM status
make backlog      # Show backlog
```

## Project Structure

```
src/rtmx/
├── cli/          # Click CLI commands
├── adapters/     # External service integrations (GitHub, Jira, MCP)
├── pytest/       # Pytest plugin for requirement markers
├── models.py     # Requirement, RTMDatabase
├── config.py     # RTMXConfig, load_config()
├── graph.py      # Dependency analysis (cycles, critical path)
├── validation.py # Database validation
└── schema.py     # Schema definitions (core, phoenix)
```

## Versioning

We use [Semantic Versioning](https://semver.org/) with strict adherence:

- **MAJOR** (x.0.0): Breaking API changes
- **MINOR** (0.x.0): New features, backward compatible
- **PATCH** (0.0.x): Bug fixes, backward compatible

### Release Process

1. Update version in `pyproject.toml`
2. Update `CHANGELOG.md` with release notes
3. Commit: `git commit -m "chore: Bump version to vX.Y.Z"`
4. Tag: `git tag vX.Y.Z`
5. Push: `git push origin main --tags`

The release workflow automatically publishes to PyPI on version tags.

### Version Constraints

- Pre-1.0: API is unstable, minor versions may break compatibility
- Post-1.0: Strict semver, deprecation warnings before removal

## Test-Driven Development (TDD)

**All code changes MUST follow TDD:**

1. **Red**: Write a failing test first
2. **Green**: Write minimal code to pass the test
3. **Refactor**: Clean up while keeping tests green

### Test Requirements

Every test function MUST have requirement markers:

```python
@pytest.mark.req("REQ-XX-NNN")      # Link to requirement
@pytest.mark.scope_unit             # or scope_integration, scope_system
@pytest.mark.technique_nominal      # or parametric, monte_carlo, stress
@pytest.mark.env_simulation         # or hil, anechoic, field
def test_feature():
    pass
```

### Test File Structure

```
tests/
├── test_models.py          # Unit tests for models
├── test_graph.py           # Unit tests for graph algorithms
├── test_validation.py      # Unit tests for validation
├── test_lifecycle_e2e.py   # E2E lifecycle tests
└── fixtures/               # Test data
```

### Running Tests

```bash
pytest tests/ -v                    # All tests
pytest tests/test_models.py -v      # Single module
pytest -k "test_status" -v          # Pattern match
pytest --cov=rtmx --cov-report=html # With coverage
```

## Behavior-Driven Development (BDD)

For feature development, write requirements BEFORE implementation:

### 1. Define Requirement

Add to `docs/rtm_database.csv`:

```csv
REQ-FEAT-001,FEATURES,API,System shall export RTM as JSON,JSON schema v1,tests/test_export.py,test_json_export,Unit Test,MISSING,MEDIUM,2,API endpoint for JSON export,0.5,REQ-CORE-001,,developer,v0.2,,,docs/requirements/FEATURES/REQ-FEAT-001.md
```

### 2. Write Specification

Create `docs/requirements/FEATURES/REQ-FEAT-001.md`:

```markdown
# REQ-FEAT-001: JSON Export

## Description
System shall export the RTM database as JSON.

## Acceptance Criteria
- [ ] Exports all requirements as JSON array
- [ ] Includes all schema fields
- [ ] Validates against JSON schema
- [ ] Handles empty database

## Test Cases
1. Export populated database
2. Export empty database
3. Export with special characters
```

### 3. Write Failing Test

```python
@pytest.mark.req("REQ-FEAT-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_json_export():
    db = RTMDatabase.load("fixtures/sample.csv")
    result = db.to_json()
    assert isinstance(result, str)
    data = json.loads(result)
    assert len(data) == len(db.requirements)
```

### 4. Implement Feature

Write minimal code to pass the test, then refactor.

### 5. Update Status

```bash
rtmx from-tests --update  # Sync test info to RTM
```

## Code Style

- **Formatter**: ruff format
- **Linter**: ruff check
- **Type checker**: mypy --strict
- **Docstrings**: Google style

### Pre-commit

```bash
make pre-commit-install  # Setup hooks
make pre-commit-run      # Run manually
```

## Adding New Features

1. Check `make backlog` for prioritized requirements
2. Create requirement spec in `docs/requirements/`
3. Write failing tests with proper markers
4. Implement feature
5. Run `make test && make lint`
6. Update CHANGELOG.md
7. Submit PR

## CLI Development

Commands are in `src/rtmx/cli/`. Each command:

1. Has its own module (e.g., `status.py`)
2. Is registered in `main.py`
3. Uses Click decorators
4. Has corresponding tests

```python
# src/rtmx/cli/newcmd.py
def run_newcmd(arg: str) -> None:
    """Implementation."""
    pass

# src/rtmx/cli/main.py
@main.command()
@click.argument("arg")
def newcmd(arg: str) -> None:
    """Command docstring."""
    from rtmx.cli.newcmd import run_newcmd
    run_newcmd(arg)
```

## Debugging

```bash
rtmx --help                    # Show all commands
rtmx status -vvv               # Verbose output
python -m rtmx status          # Run as module
pytest -v -s --tb=long         # Verbose test output
```

## Architecture Decisions

- **CSV over SQLite**: Human-readable, git-friendly, AI-parseable
- **Click over argparse**: Better UX, composable commands
- **Pydantic-style validation**: Type safety without runtime overhead
- **Lazy imports in CLI**: Fast startup time
- When you push, always monitor CI in a background process. Fix pipeline errors immediately. A broken pipeline always becomes the team's highest priority.
- The contact information for the rtmx project is ioTACTICAL Engineering, engineering@iotactical.co.
- The company name is ioTACTICAL LLC. The contact information is "ioTACTICAL Engineering" reachable at engineering@iotactical.co.
- Never use --no-verify or SKIP= when we run a commit. Pre-commit checks exist to fully verify what we are pushing to main. Catching an error in main that should have been caught locally is a quality escape, and we should always minimize or eliminate quality escapes. Quality lives as far left as possible, and for this project, that's in the local development environment.

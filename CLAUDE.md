# CLAUDE.md

This file provides guidance to Claude Code when working with the RTMX codebase.

## Repository Ecosystem

This is the core RTMX CLI client, part of the multi-repo system:

| Repo | Purpose | Relationship |
|------|---------|--------------|
| rtmx.ai | Website & docs | Has rtmx as submodule |
| **rtmx** (this) | CLI client | Core library |
| rtmx-sync | Real-time coordination | Imports rtmx>=0.0.5 |

When working across repos:
- API changes here → update rtmx-sync dependency
- Doc changes here → update rtmx.ai submodule
- Breaking changes → coordinate across all repos

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

## Gherkin Feature Specifications

RTMX uses pytest-bdd for executable Gherkin specifications. Feature files live in `features/` and are language-agnostic.

### Directory Structure

```
features/                    # Gherkin feature files (portable)
├── cli/                     # CLI command features
│   ├── status.feature
│   └── backlog.feature
└── sync/                    # Collaboration features
    ├── collaboration.feature
    └── offline.feature

tests/bdd/                   # Python step definitions
├── conftest.py              # BDD fixtures
├── steps/
│   ├── common_steps.py      # Shared Given/When/Then
│   └── cli_steps.py         # CLI-specific steps
└── scenarios/               # pytest-bdd test modules
    └── test_cli_status.py
```

### Writing Feature Files

Every feature file MUST:
1. Link to requirements via `@REQ-XXX` tags
2. Include test scope and technique tags
3. Follow Gherkin best practices (Background for setup, Scenario Outline for data-driven)

```gherkin
@REQ-CLI-001 @REQ-UX-001 @cli
Feature: RTM Status Display
  As a developer using RTMX
  I want to see the current RTM completion status
  So that I can track project progress

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Display status summary
    Given the RTM database has 10 requirements
    And 5 requirements are COMPLETE
    When I run "rtmx status"
    Then the command should succeed
    And I should see "50%" in the output
```

### Tag Conventions

| Tag Pattern | Purpose | Example |
|-------------|---------|---------|
| `@REQ-XXX-NNN` | Link to requirement | `@REQ-CLI-001` |
| `@scope_*` | Test scope | `@scope_system` |
| `@technique_*` | Test technique | `@technique_nominal` |
| `@cli/@sync/@web` | Component | `@cli` |
| `@phase-N` | Development phase | `@phase-10` |

### Step Definition Patterns

Use universal patterns that translate across languages:

```python
# tests/bdd/steps/common_steps.py
from pytest_bdd import given, when, then, parsers

@given("an initialized RTMX project")
def initialized_project(tmp_path):
    """Create project with rtmx.yaml and database."""
    ...

@when(parsers.parse('I run "{command}"'))
def run_command(context, command):
    """Execute CLI command. Universal pattern for any language."""
    ...

@then("the command should succeed")
def command_succeeds(context):
    assert context["result"].returncode == 0
```

### Running BDD Tests

```bash
pytest tests/bdd/ -v                      # Run BDD tests only
pytest tests/ -v                          # Run all tests (unit + BDD)
pytest tests/bdd/ -v --gherkin-terminal-reporter  # Verbose Gherkin output
```

### BDD Workflow for New Features

1. **Write Feature Spec First**: Create `.feature` file with scenarios
2. **Add Requirement Tags**: Link to existing or new requirements
3. **Write Step Definitions**: Implement Given/When/Then in Python
4. **Create Scenario Runner**: Add `tests/bdd/scenarios/test_*.py`
5. **Run and Iterate**: Use failing scenarios to drive implementation

### Multi-Language Portability

Feature files are portable across languages. When adding SDKs in other languages:
- Keep same `.feature` files
- Write new step definitions in target language
- Use language-appropriate BDD runner (cucumber-js, godog, etc.)

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

## Parallel Development with Git Worktrees

When multiple requirements are **mutually exclusive** (no shared file dependencies, no blocking relationships), use parallel Claude Code agents on separate Git worktrees to accelerate development.

### When to Parallelize

Requirements are safe to parallelize when:
- They touch **different files** (check spec's "Files to Modify" section)
- Neither **blocks** the other (check `Blocks` / `Dependencies` in specs)
- They belong to **different components** (e.g., CLI vs adapters vs tests)

### Worktree Setup

```bash
# Create worktrees for parallel work
git worktree add ../rtmx-feat-a -b feat/REQ-XXX-001
git worktree add ../rtmx-feat-b -b feat/REQ-YYY-001

# Each agent works in its own worktree
cd ../rtmx-feat-a  # Agent 1
cd ../rtmx-feat-b  # Agent 2
```

### Workflow

1. **Identify candidates**: Run `make backlog` and find unblocked requirements
2. **Verify independence**: Read specs to confirm no file overlap
3. **Create worktrees**: One per requirement, with feature branches
4. **Spawn agents**: Each agent implements one requirement in its worktree
5. **Merge sequentially**: After agents complete, merge branches to main one at a time
6. **Clean up**: `git worktree remove ../rtmx-feat-a`

### Example: Parallel Phase Work

```bash
# REQ-MCP-001 (adapters/mcp/) and REQ-GIT-002 (git hooks) are independent
git worktree add ../rtmx-mcp -b feat/REQ-MCP-001
git worktree add ../rtmx-git -b feat/REQ-GIT-002

# Agent 1: Implement MCP spec reading in ../rtmx-mcp
# Agent 2: Implement pre-commit hooks in ../rtmx-git

# After both complete:
git merge feat/REQ-MCP-001
git merge feat/REQ-GIT-002
git worktree remove ../rtmx-mcp
git worktree remove ../rtmx-git
```

### Caution

- **Never parallelize blocking requirements** - check `blockedBy` fields
- **Avoid shared files** - if two reqs modify `models.py`, work sequentially
- **Sync database changes** - `docs/rtm_database.csv` edits should be coordinated

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
- The contact information for the rtmx project is RTMX Engineering, dev@rtmx.ai.
- The company name is ioTACTICAL LLC (owner of RTMX). Technical support: dev@rtmx.ai, Sales: sales@rtmx.ai, Help: help@rtmx.ai.
- Never use --no-verify or SKIP= when we run a commit. Pre-commit checks exist to fully verify what we are pushing to main. Catching an error in main that should have been caught locally is a quality escape, and we should always minimize or eliminate quality escapes. Quality lives as far left as possible, and for this project, that's in the local development environment.
- Every tag for release to pypi should be accompanied with a thorough documentation review and update.

# REQ-BDD-INF-001: BDD Infrastructure

## Status: COMPLETE
## Priority: HIGH
## Phase: 10 (Prerequisite)

## Description
RTMX shall have BDD (Behavior-Driven Development) infrastructure using pytest-bdd framework, enabling Gherkin feature specifications that trace to requirements and run alongside existing pytest unit tests.

## Rationale
BDD provides executable specifications that serve as living documentation. By implementing BDD infrastructure before Phase 10 (Collaboration), we can write feature specs for Sync functionality before implementation, ensuring clear acceptance criteria and testable requirements.

## Acceptance Criteria
- [x] pytest-bdd dependency added to pyproject.toml
- [x] `features/` directory created with CLI and Sync subdirectories
- [x] `tests/bdd/` directory created with step definitions
- [x] Common step definitions implemented (Given/When/Then patterns)
- [x] First CLI feature file created (`features/cli/status.feature`)
- [x] BDD tests run with standard `pytest` command
- [x] Requirement tags (`@REQ-XXX`) link features to RTM database
- [x] pytest markers (scope, technique, env) work with BDD scenarios

## Technical Notes

### Framework Selection: pytest-bdd
- Native pytest integration (existing markers work)
- Unified test runner (`pytest` runs BDD + unit tests together)
- Fixture reuse from existing `conftest.py`
- CI integration unchanged
- Coverage reporting included

### Dependencies
```toml
[project.optional-dependencies]
bdd = ["pytest-bdd>=7.0", "gherkin-official>=24.0"]
```

### Directory Structure
```
rtmx/
├── features/                    # Gherkin feature files (portable)
│   ├── cli/                     # CLI command features
│   │   ├── status.feature
│   │   └── backlog.feature
│   └── sync/                    # Phase 10 features
│       ├── collaboration.feature
│       └── offline.feature
└── tests/bdd/                   # Step definitions (Python)
    ├── conftest.py              # BDD fixtures
    ├── steps/
    │   ├── common_steps.py      # Shared Given/When/Then
    │   └── cli_steps.py         # CLI-specific steps
    └── scenarios/               # pytest-bdd test modules
        └── test_cli_status.py
```

### Requirement Tracing Convention
```gherkin
@REQ-CLI-001 @REQ-UX-001 @cli
Feature: RTM Status Display
  As a developer using RTMX
  I want to see the current RTM completion status

  @scope_system @technique_nominal
  Scenario: Display status summary
    Given an initialized RTMX project with 5 requirements
    When I run "rtmx status"
    Then I should see the completion percentage
```

### Tag Conventions
- `@REQ-XXX-NNN` - Links to requirement in RTM database
- `@scope_unit/@scope_integration/@scope_system` - Test scope
- `@technique_nominal/@technique_stress` - Test technique
- `@cli/@sync/@web` - Component tags

## Multi-Language Portability
Feature files are language-agnostic (`.feature` is plain text). When RTMX adds client SDKs in other languages (JS/TS, Go, Rust), the same feature files can be used with language-specific step definitions:
- Python: pytest-bdd
- JavaScript: cucumber-js
- Go: godog
- Rust: cucumber-rs

## Test Cases
- `tests/test_bdd.py::TestBDDInfrastructure::test_pytest_bdd_installed`
- `tests/test_bdd.py::TestBDDInfrastructure::test_feature_files_exist`
- `tests/test_bdd.py::TestBDDInfrastructure::test_bdd_tests_run_with_pytest`
- `tests/bdd/scenarios/test_cli_status.py::test_display_status_summary`

## Dependencies
- REQ-PYTEST-001: Pytest plugin (markers infrastructure)

## Blocks
- REQ-COLLAB-001: Sync server (Feature specs define acceptance criteria)
- REQ-CRDT-007: Y.Text collaborative editing

## Implementation Checklist
1. Add pytest-bdd to pyproject.toml
2. Create `features/` directory structure
3. Create `tests/bdd/` directory with conftest.py
4. Write `tests/bdd/steps/common_steps.py`
5. Write `features/cli/status.feature`
6. Write `tests/bdd/scenarios/test_cli_status.py`
7. Verify `pytest tests/bdd/ -v` runs successfully
8. Update CLAUDE.md with BDD directives

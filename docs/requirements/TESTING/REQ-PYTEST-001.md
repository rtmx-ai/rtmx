# REQ-PYTEST-001: Minimal Pytest Plugin

## Status: MISSING
## Priority: HIGH
## Phase: 13

## Description
The Python `rtmx` package shall be reduced to a **minimal pytest plugin** that registers requirement markers and outputs JSON results for import by the Go CLI. This replaces the current full-featured Python implementation after Go CLI deprecation.

## Rationale
After REQ-DIST-002 (Go CLI) reaches feature parity:
- Python CLI becomes redundant (Go handles all commands)
- Maintaining two implementations is unsustainable
- pytest plugin must remain Python (runs inside pytest process)
- Minimal plugin reduces maintenance burden to essential functionality

## Scope

### What This Plugin Does
1. **Marker Registration**: Registers `@pytest.mark.req`, `@pytest.mark.scope_*`, `@pytest.mark.technique_*`, `@pytest.mark.env_*`
2. **Result Collection**: Captures test outcomes with requirement associations
3. **JSON Output**: Writes `rtmx-results.json` for Go CLI import
4. **Marker Discovery**: Scans source files for markers (retained from current implementation)

### What This Plugin Does NOT Do
- CLI commands (use Go CLI)
- Database parsing/writing (use Go CLI)
- Graph algorithms (use Go CLI)
- Validation (use Go CLI)
- Sync/adapters (use Go CLI)

## Acceptance Criteria

### Core Functionality
- [ ] `pip install rtmx` installs minimal package (<500KB)
- [ ] Zero runtime dependencies beyond pytest
- [ ] Markers registered automatically via pytest plugin entry point
- [ ] `@pytest.mark.req("REQ-XXX-NNN")` works unchanged from current implementation
- [ ] `@pytest.mark.scope_unit`, `scope_integration`, `scope_system`, `scope_acceptance` work
- [ ] `@pytest.mark.technique_nominal`, `technique_parametric`, etc. work
- [ ] `@pytest.mark.env_simulation`, `env_hil`, etc. work

### JSON Output
- [ ] `--rtmx-output=results.json` pytest option writes results file
- [ ] JSON schema matches RTMX marker specification (REQ-LANG-007)
- [ ] Output includes: req_id, test_name, test_module, outcome, duration, markers
- [ ] Handles parametrized tests with requirement inheritance
- [ ] Handles class-level markers applied to all methods

### Integration with Go CLI
- [ ] `rtmx from-pytest results.json` imports results (Go CLI command)
- [ ] Round-trip: pytest → JSON → Go CLI → database update
- [ ] Error handling for malformed JSON or missing fields

### Migration
- [ ] Deprecation warnings for removed features (e.g., `rtmx status` Python command)
- [ ] Clear error messages pointing to Go CLI for deprecated commands
- [ ] Migration guide in documentation

## Package Structure (Post-Deprecation)

```
rtmx/                           # Minimal package
├── __init__.py                 # Version only
├── pytest/
│   ├── __init__.py
│   ├── plugin.py               # Pytest plugin hooks
│   └── reporter.py             # JSON result reporter
└── markers/                    # Marker discovery (retained)
    ├── __init__.py
    ├── models.py               # MarkerInfo dataclass
    ├── detection.py            # Language detection
    ├── discover.py             # File scanning
    ├── schema.py               # JSON schema
    └── parsers/
        ├── __init__.py
        ├── python.py           # Python parser
        └── comment.py          # Generic comment parser
```

### Files Removed (Post-Deprecation)

```
# All of these move to Go CLI
src/rtmx/cli/                   # REMOVED - Go CLI
src/rtmx/models.py              # REMOVED - Go CLI
src/rtmx/graph.py               # REMOVED - Go CLI
src/rtmx/validation.py          # REMOVED - Go CLI
src/rtmx/config.py              # REMOVED - Go CLI
src/rtmx/sync/                  # REMOVED - Go CLI
src/rtmx/adapters/              # REMOVED - Go CLI
src/rtmx/web/                   # REMOVED - Go CLI
src/rtmx/auth/                  # REMOVED - Go CLI
src/rtmx/ziti/                  # REMOVED - Go CLI
```

## Technical Implementation

### Plugin Entry Point (pyproject.toml)

```toml
[project.entry-points.pytest11]
rtmx = "rtmx.pytest.plugin"
```

### Marker Registration

```python
# rtmx/pytest/plugin.py
def pytest_configure(config):
    config.addinivalue_line("markers", "req(id): Link test to requirement ID")
    config.addinivalue_line("markers", "scope_unit: Unit test scope")
    config.addinivalue_line("markers", "scope_integration: Integration test scope")
    # ... etc
```

### Result Collection

```python
# rtmx/pytest/reporter.py
class RtmxReporter:
    def __init__(self, output_path: Path):
        self.output_path = output_path
        self.results: list[dict] = []

    def pytest_runtest_logreport(self, report):
        if report.when == "call":
            markers = self._extract_markers(report)
            self.results.append({
                "test_name": report.nodeid,
                "outcome": report.outcome,
                "duration": report.duration,
                "markers": markers,
            })

    def pytest_sessionfinish(self):
        with open(self.output_path, "w") as f:
            json.dump({"results": self.results}, f, indent=2)
```

### JSON Output Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "results": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["test_name", "outcome"],
        "properties": {
          "test_name": { "type": "string" },
          "outcome": { "enum": ["passed", "failed", "skipped", "xfailed", "xpassed"] },
          "duration": { "type": "number" },
          "markers": {
            "type": "object",
            "properties": {
              "req_id": { "type": "string", "pattern": "^REQ-[A-Z]+-[0-9]+$" },
              "scope": { "enum": ["unit", "integration", "system", "acceptance"] },
              "technique": { "enum": ["nominal", "parametric", "monte_carlo", "stress", "boundary"] },
              "env": { "enum": ["simulation", "hil", "anechoic", "field"] }
            }
          }
        }
      }
    },
    "metadata": {
      "type": "object",
      "properties": {
        "timestamp": { "type": "string", "format": "date-time" },
        "pytest_version": { "type": "string" },
        "rtmx_version": { "type": "string" }
      }
    }
  }
}
```

## Test Cases
1. `tests/test_pytest_plugin.py::test_marker_registration` - Markers registered on configure
2. `tests/test_pytest_plugin.py::test_req_marker_extraction` - req marker captured
3. `tests/test_pytest_plugin.py::test_scope_markers` - scope markers captured
4. `tests/test_pytest_plugin.py::test_json_output_format` - JSON matches schema
5. `tests/test_pytest_plugin.py::test_parametrized_tests` - Parametrized handling
6. `tests/test_pytest_plugin.py::test_class_level_markers` - Class marker inheritance
7. `tests/test_pytest_plugin.py::test_minimal_dependencies` - No extra deps
8. `tests/test_pytest_plugin.py::test_deprecation_warnings` - Deprecated features warn

## Dependencies
- REQ-DIST-002: Go CLI (for `rtmx from-pytest` command)
- REQ-LANG-007: Marker specification (JSON schema)

## Blocks
- None (leaf requirement)

## Effort
2.0 weeks

## Migration Timeline

| Milestone | Action |
|-----------|--------|
| Go CLI v1.0 | Feature parity achieved |
| rtmx v1.0 | Deprecation warnings added to Python CLI |
| rtmx v1.1 | Python CLI commands removed |
| rtmx v2.0 | Minimal plugin only, major version bump |

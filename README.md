# RTMX

Requirements Traceability Matrix toolkit for GenAI-driven development.

[![PyPI version](https://img.shields.io/pypi/v/rtmx.svg)](https://pypi.org/project/rtmx/)
[![Python versions](https://img.shields.io/pypi/pyversions/rtmx.svg)](https://pypi.org/project/rtmx/)
[![License](https://img.shields.io/github/license/iotactical/rtm.svg)](https://github.com/iotactical/rtm/blob/main/LICENSE)

## Overview

RTMX provides tools for managing requirements traceability in software projects, with focus on:

- **GenAI Integration**: Token-efficient data model for AI-driven development
- **Compliance Frameworks**: Support for CMMC, FedRAMP, and other compliance standards
- **Pytest Integration**: Link tests to requirements with markers
- **Dependency Analysis**: Cycle detection, critical path analysis

## Installation

```bash
pip install rtmx
```

For development:

```bash
pip install rtmx[dev]
```

## Quick Start

### Initialize RTM in your project

```bash
rtmx init
```

This creates:
- `docs/rtm_database.csv` - Requirements database
- `docs/requirements/` - Requirement specification files
- `rtmx.yaml` - Configuration file

### Check status

```bash
rtmx status           # Summary
rtmx status -v        # Category breakdown
rtmx status -vv       # Subcategory breakdown
rtmx status -vvv      # All requirements
```

### View backlog

```bash
rtmx backlog              # All incomplete
rtmx backlog --phase 1    # Phase 1 only
rtmx backlog --critical   # Critical path only
```

### Check for issues

```bash
rtmx cycles               # Detect circular dependencies
rtmx reconcile            # Check dependency reciprocity
rtmx reconcile --execute  # Fix reciprocity issues
```

## Python API

```python
from rtmx import RTMDatabase, Status

# Load database
db = RTMDatabase.load("docs/rtm_database.csv")

# Query requirements
req = db.get("REQ-SW-001")
incomplete = db.filter(status=Status.MISSING)
phase1 = db.filter(phase=1)

# Graph operations
cycles = db.find_cycles()
blockers = db.critical_path()

# Validation
errors = db.validate()
violations = db.check_reciprocity()

# Modify and save
db.update("REQ-SW-001", status=Status.COMPLETE)
db.save()
```

## Pytest Integration

RTMX provides pytest markers for requirement traceability:

```python
import pytest

@pytest.mark.req("REQ-SW-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_feature_x():
    """Test that validates REQ-SW-001."""
    assert feature_x() == expected
```

Available markers:

- `req(id)` - Link test to requirement
- Scope: `scope_unit`, `scope_integration`, `scope_system`
- Technique: `technique_nominal`, `technique_parametric`, `technique_monte_carlo`, `technique_stress`
- Environment: `env_simulation`, `env_hil`, `env_anechoic`, `env_static_field`, `env_dynamic_field`

## Configuration

Create `rtmx.yaml` in your project root:

```yaml
rtmx:
  database: docs/rtm_database.csv
  requirements_dir: docs/requirements
  schema: core  # or "phoenix" for extended schema
  pytest:
    marker_prefix: "req"
    register_markers: true
```

## RTM Schema

### Core Schema (20 columns)

| Column | Type | Required | Description |
|--------|------|----------|-------------|
| `req_id` | string | Yes | Unique identifier (e.g., REQ-SW-001) |
| `category` | string | Yes | High-level grouping |
| `subcategory` | string | No | Detailed classification |
| `requirement_text` | string | Yes | Human-readable description |
| `target_value` | string | No | Quantitative criteria |
| `test_module` | string | No | Test file path |
| `test_function` | string | No | Test function name |
| `validation_method` | string | No | Testing approach |
| `status` | string | Yes | COMPLETE/PARTIAL/MISSING |
| `priority` | string | No | P0/HIGH/MEDIUM/LOW |
| `phase` | integer | No | Development phase |
| `notes` | string | No | Additional context |
| `effort_weeks` | float | No | Estimated effort |
| `dependencies` | list | No | Pipe-separated req IDs |
| `blocks` | list | No | Pipe-separated req IDs |
| `assignee` | string | No | Owner |
| `sprint` | string | No | Target version |
| `started_date` | date | No | YYYY-MM-DD |
| `completed_date` | date | No | YYYY-MM-DD |
| `requirement_file` | string | No | Path to spec file |

### Phoenix Extension

The Phoenix schema adds validation taxonomy columns:

- Scope: `scope_unit`, `scope_integration`, `scope_system`
- Technique: `technique_nominal`, `technique_parametric`, etc.
- Environment: `env_simulation`, `env_hil`, `env_anechoic`, etc.
- Metrics: `baseline_metric`, `current_metric`, `target_metric`

## Makefile Integration

Add to your Makefile:

```makefile
.PHONY: rtm rtm-v rtm-vv rtm-vvv backlog

rtm:
	rtmx status

rtm-v:
	rtmx status -v

rtm-vv:
	rtmx status -vv

rtm-vvv:
	rtmx status -vvv

backlog:
	rtmx backlog
```

## Jetstream Integration

RTMX is designed as a foundation for the Jetstream digital engineering platform:

- CMMC Level 2 compliance mapping
- FedRAMP High authorization support
- Prometheus metrics export (planned)
- Multi-project federation (planned)

## Development

```bash
# Clone and install
git clone https://github.com/iotactical/rtm.git
cd rtm
make dev

# Run tests
make test

# Run linter
make lint

# Format code
make format
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

# RTMX Schema Documentation

This document describes the RTM database schema used by RTMX.

## Overview

RTMX uses a CSV-based requirements database with a well-defined schema. The schema is extensible, supporting both a **Core Schema** (20 columns) and a **Phoenix Extension Schema** (45+ columns) for advanced validation taxonomy.

## Core Schema (20 Columns)

The core schema provides all essential fields for basic requirements traceability.

| # | Column | Type | Required | Description |
|---|--------|------|----------|-------------|
| 1 | `req_id` | string | Yes | Unique requirement identifier (e.g., REQ-SW-001) |
| 2 | `category` | string | Yes | High-level grouping (SOFTWARE, PERFORMANCE, TESTING, etc.) |
| 3 | `subcategory` | string | No | Detailed classification (ALGORITHM, ACCURACY, UNIT, etc.) |
| 4 | `requirement_text` | string | Yes | Human-readable requirement description |
| 5 | `target_value` | string | No | Quantitative acceptance criteria |
| 6 | `test_module` | string | No | Python test file path (e.g., tests/test_signal.py) |
| 7 | `test_function` | string | No | Test function name (e.g., test_range_processing) |
| 8 | `validation_method` | string | No | Testing approach (Unit Test, Integration Test, etc.) |
| 9 | `status` | enum | Yes | COMPLETE, PARTIAL, MISSING, or NOT_STARTED |
| 10 | `priority` | enum | No | P0, HIGH, MEDIUM, LOW (default: MEDIUM) |
| 11 | `phase` | integer | No | Development phase number (1, 2, 3, etc.) |
| 12 | `notes` | string | No | Additional context or implementation notes |
| 13 | `effort_weeks` | float | No | Estimated effort in weeks |
| 14 | `dependencies` | string | No | Pipe-separated list of requirement IDs this depends on |
| 15 | `blocks` | string | No | Pipe-separated list of requirements this blocks |
| 16 | `assignee` | string | No | Person responsible for implementation |
| 17 | `sprint` | string | No | Target sprint or version |
| 18 | `started_date` | date | No | When work began (YYYY-MM-DD) |
| 19 | `completed_date` | date | No | When completed (YYYY-MM-DD) |
| 20 | `requirement_file` | string | No | Path to detailed specification markdown |

## Field Details

### Status Values

| Value | Description | Completion % |
|-------|-------------|--------------|
| `COMPLETE` | Fully implemented and all tests passing | 100% |
| `PARTIAL` | Partially implemented, some tests passing | 50% |
| `MISSING` | Not started, no implementation | 0% |
| `NOT_STARTED` | Explicitly marked as not started | 0% |

### Priority Values

| Value | Description | Sort Order |
|-------|-------------|------------|
| `P0` | Critical - must be done immediately | 1 (highest) |
| `HIGH` | High priority | 2 |
| `MEDIUM` | Normal priority (default) | 3 |
| `LOW` | Low priority, nice to have | 4 (lowest) |

### Validation Method Values

Standard validation methods include:
- `Unit Test` - Single component isolation test
- `Integration Test` - Multi-component interaction test
- `System Test` - End-to-end system validation
- `Analysis` - Manual or automated analysis
- `Inspection` - Code or documentation review
- `Design` - Design-time verification

### Dependency Format

Dependencies and blocks use pipe-separated (`|`) requirement IDs:

```
dependencies: REQ-SW-001|REQ-SW-002
blocks: REQ-PERF-001|REQ-PERF-002
```

Space-separated format is also supported for backwards compatibility:

```
dependencies: REQ-SW-001 REQ-SW-002
```

## Phoenix Extension Schema

The Phoenix schema extends the core schema with 25 additional columns for advanced validation taxonomy. This is used in defense and aerospace projects requiring detailed test classification.

### Scope Markers (Boolean)

| Column | Description |
|--------|-------------|
| `scope_unit` | Single component isolation test (<1ms typical) |
| `scope_integration` | Multiple components (<100ms typical) |
| `scope_system` | End-to-end system test (<1s typical) |

### Technique Markers (Boolean)

| Column | Description |
|--------|-------------|
| `technique_nominal` | Happy path, typical parameters |
| `technique_parametric` | Systematic parameter space exploration |
| `technique_monte_carlo` | Random scenario testing (1-10s) |
| `technique_stress` | Boundary/edge cases, extreme conditions |

### Environment Markers (Boolean)

| Column | Description |
|--------|-------------|
| `env_simulation` | Pure software, synthetic signals |
| `env_hil` | Real hardware, controlled signals |
| `env_anechoic` | RF characterization chamber |
| `env_static_field` | Outdoor, stationary targets |
| `env_dynamic_field` | Outdoor, moving targets |

### Metrics Columns

| Column | Type | Description |
|--------|------|-------------|
| `baseline_metric` | float | Initial measurement value |
| `current_metric` | float | Current measurement value |
| `target_metric` | float | Target measurement value |
| `metric_unit` | string | Unit of measurement |

### Hardware Columns

| Column | Type | Description |
|--------|------|-------------|
| `lead_time_weeks` | float | Hardware procurement lead time |
| `supplier_part` | string | Supplier part number or identifier |

## Example CSV

```csv
req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-SW-001,SOFTWARE,ALGORITHM,Implement range processing algorithm,Range resolution <=1m,tests/test_signal.py,test_range_processing,Unit Test,COMPLETE,HIGH,1,Core DSP algorithm,2.0,,,alice,v0.1,2025-01-15,2025-01-29,docs/requirements/SOFTWARE/REQ-SW-001.md
REQ-SW-002,SOFTWARE,ALGORITHM,Implement Doppler processing,Velocity resolution <=0.5m/s,tests/test_signal.py,test_doppler_processing,Unit Test,PARTIAL,HIGH,1,FFT-based processing,1.5,REQ-SW-001,,bob,v0.1,2025-01-20,,docs/requirements/SOFTWARE/REQ-SW-002.md
```

## Schema Selection

Configure the schema in `rtmx.yaml`:

```yaml
rtmx:
  schema: core      # Use 20-column core schema (default)
  # schema: phoenix # Use extended Phoenix schema
```

## Custom Extensions

The schema is extensible. Additional columns beyond the defined schema are preserved in the `extra` field when loading requirements:

```python
from rtmx import RTMDatabase

db = RTMDatabase.load("docs/rtm_database.csv")
req = db.get("REQ-SW-001")
custom_value = req.extra.get("custom_column")
```

## Validation

RTMX validates the database against the schema:

```bash
rtmx health                    # Run all validation checks
rtmx config --validate         # Validate configuration
```

Validation checks include:
- Required fields are present and non-empty
- Status values are valid enum members
- Priority values are valid enum members
- Phase values are valid integers
- No duplicate requirement IDs
- Dependencies reference existing requirements
- Blocks reference existing requirements
- Reciprocity: if A blocks B, then B depends on A

## API Reference

```python
from rtmx.schema import CORE_SCHEMA, PHOENIX_SCHEMA, Schema, Column

# Get column definitions
for col in CORE_SCHEMA.columns:
    print(f"{col.name}: {col.type} (required={col.required})")

# Extend schema
from rtmx.schema import register_schema

custom_schema = CORE_SCHEMA.extend([
    Column("custom_field", ColumnType.STRING, required=False),
])
register_schema(custom_schema)
```

## See Also

- [LIFECYCLE.md](LIFECYCLE.md) - Full RTMX lifecycle documentation
- [README.md](../README.md) - Quick start guide
- [rtmx.yaml Configuration](configuration.md) - Configuration options

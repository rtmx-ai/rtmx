# REQ-PLUGIN-006: Built-in Domain Schema Plugins

## Metadata
- **Category**: PLUGIN
- **Subcategory**: Schema
- **Priority**: HIGH
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: REQ-PLUGIN-005
- **Blocks**: (none)

## Requirement

RTMX shall ship built-in schema plugins for common certification domains:
Phoenix (defense/aerospace test taxonomy, already documented), DO-178C
(airborne software), and ISO 26262 (automotive functional safety). These
ship as YAML files in the rtmx distribution and are available without
additional installation.

## Acceptance Criteria

1. `rtmx schema list` shows phoenix, do178c, and iso26262 as built-in
2. Each schema defines typed columns with validation rules
3. `rtmx init --schema phoenix` creates a database with Phoenix columns
4. Health checks enforce domain-specific rules (e.g., DAL coverage)
5. Schema documentation generates from the YAML definitions

## Files to Create

- `schemas/phoenix.yaml` -- Defense/aerospace test taxonomy (from docs/schema.md)
- `schemas/do178c.yaml` -- Airborne software certification
- `schemas/iso26262.yaml` -- Automotive functional safety

## Effort Estimate

1 week (schema YAML definitions + validation rules + tests)

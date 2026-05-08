# REQ-PLUGIN-005c: Schema registry and config-driven selection

## Metadata
- **Category**: PLUGIN
- **Subcategory**: Schema
- **Priority**: HIGH
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: REQ-PLUGIN-005b
- **Blocks**: REQ-PLUGIN-005d

## Requirement

Create a schema registry that maps schema names (core, phoenix) to
Schema instances. Wire into config so rtmx.schema selects the active
schema. Database load validates against the active schema.

## Acceptance Criteria

1. Registry maps schema name string to Schema instance
2. Register() adds a schema, Get() retrieves by name
3. Config.RTMX.Schema selects active schema (default: core)
4. database.Load() validates header against active schema when available
5. Unknown schema name produces a clear error

## Effort Estimate

0.5 weeks

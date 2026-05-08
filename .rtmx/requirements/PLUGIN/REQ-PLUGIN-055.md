# REQ-PLUGIN-055: Schema extension API for custom columns

## Metadata
- **Category**: PLUGIN
- **Subcategory**: Schema
- **Priority**: MEDIUM
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-PLUGIN-052
- **Blocks**: REQ-PLUGIN-007

## Requirement

Schema.Extend() returns a new Schema with additional columns appended.
Custom schemas can be defined in .rtmx/schema.yaml and loaded at
startup. Extra columns in the database are validated against the
custom schema rather than silently stored in Extra.

## Acceptance Criteria

1. Schema.Extend([]Column) returns new schema with appended columns
2. .rtmx/schema.yaml defines custom columns in YAML
3. Custom columns loaded and registered at config load time
4. database.Load() validates custom columns when schema is active
5. Extra columns matching custom schema are type-checked

## Effort Estimate

0.5 weeks

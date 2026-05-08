# REQ-PLUGIN-005e: Phoenix extension schema definition

## Metadata
- **Category**: PLUGIN
- **Subcategory**: Schema
- **Priority**: HIGH
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: REQ-PLUGIN-005c
- **Blocks**: REQ-PLUGIN-006

## Requirement

Define the Phoenix extension schema as a Schema instance registered
under name "phoenix". Extends core schema with scope/technique/env
boolean columns, metrics columns, and hardware columns per docs/schema.md.

## Acceptance Criteria

1. PhoenixSchema extends CoreSchema with 25+ additional columns
2. Scope columns (scope_unit, scope_integration, scope_system) typed as bool
3. Technique/env columns typed as bool
4. Metrics columns typed as float
5. Test validates PhoenixSchema column count and types
6. rtmx health with schema: phoenix validates Phoenix-format databases

## Effort Estimate

0.25 weeks

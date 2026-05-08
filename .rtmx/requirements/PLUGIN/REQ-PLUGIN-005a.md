# REQ-PLUGIN-005a: Schema type with column definitions

## Metadata
- **Category**: PLUGIN
- **Subcategory**: Schema
- **Priority**: P0
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: REQ-GO-008
- **Blocks**: REQ-PLUGIN-005b

## Requirement

Create internal/schema/ package with Schema type containing Column
definitions (name, type, required flag, description, enum values).
Define ColumnType enum (string, int, float, bool, date, enum, set).

## Acceptance Criteria

1. Schema struct holds ordered list of Column definitions
2. Column struct has Name, Type, Required, Description, EnumValues
3. ColumnType enum covers string, int, float, bool, date, enum, set
4. Schema.Validate(row) validates a CSV row against column definitions
5. Table-driven tests for all column types

## Effort Estimate

0.5 weeks

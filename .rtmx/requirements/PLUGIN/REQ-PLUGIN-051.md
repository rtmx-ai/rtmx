# REQ-PLUGIN-051: Built-in core schema definition

## Metadata
- **Category**: PLUGIN
- **Subcategory**: Schema
- **Priority**: P0
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: REQ-PLUGIN-050
- **Blocks**: REQ-PLUGIN-052

## Requirement

Define the Core schema (21 standard columns) as a Schema instance
using the Column type from REQ-PLUGIN-050. The schema must match
the existing standardColumns in internal/database/csv.go.

## Acceptance Criteria

1. CoreSchema variable defines all 21 columns with correct types
2. Column types match: req_id=string, phase=int, effort_weeks=float, etc.
3. Required columns (req_id, category, requirement_text) marked required
4. Test validates CoreSchema matches standardColumns list
5. Test validates CoreSchema.Validate() accepts a valid row

## Effort Estimate

0.25 weeks

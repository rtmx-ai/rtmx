# REQ-PLUGIN-005d: Schema health checks in rtmx health

## Metadata
- **Category**: PLUGIN
- **Subcategory**: Schema
- **Priority**: HIGH
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: REQ-PLUGIN-005c
- **Blocks**: REQ-PLUGIN-005e

## Requirement

Add schema validation check to rtmx health that reports whether the
database conforms to the active schema: missing columns, extra columns,
type violations.

## Acceptance Criteria

1. rtmx health includes "Schema conformance" check
2. Reports PASS when database matches active schema
3. Reports WARN when extra columns exist (extensible)
4. Reports FAIL when required columns are missing
5. JSON output includes schema check details

## Effort Estimate

0.5 weeks

# REQ-PM-007: Custom fields schema

## Status: NOT STARTED
## Priority: MEDIUM
## Phase: 17
## Estimated Effort: 2.5 weeks

## Description

System shall support user-defined custom fields with type validation. Teams can extend the RTM schema with project-specific fields configured in `rtmx.yaml`. Custom fields support various data types and are included in all export formats.

## Acceptance Criteria

- [ ] Custom fields defined in `rtmx.yaml` under `custom_fields` key
- [ ] Supported field types: `string`, `int`, `float`, `bool`, `date`, `enum`, `list`
- [ ] `rtmx field set <req_id> <field_name> <value>` sets custom field value
- [ ] `rtmx field get <req_id> <field_name>` retrieves custom field value
- [ ] Type validation: values must match declared field type
- [ ] Enum fields validate against allowed values list
- [ ] Date fields support ISO 8601 format with validation
- [ ] List fields support comma-separated values
- [ ] Custom fields appear in `rtmx status` output (configurable)
- [ ] Custom fields included in JSON/CSV exports
- [ ] `rtmx field list` shows all defined custom fields with types
- [ ] Required custom fields enforced during requirement creation
- [ ] Default values supported for optional custom fields
- [ ] Field migration: adding new field populates existing requirements with default
- [ ] `rtmx validate` checks all custom field values against schema

## Test Cases

- `tests/test_custom_fields.py::test_define_string_field` - Define string custom field
- `tests/test_custom_fields.py::test_define_enum_field` - Define enum field with values
- `tests/test_custom_fields.py::test_set_field_value` - Set custom field value
- `tests/test_custom_fields.py::test_get_field_value` - Get custom field value
- `tests/test_custom_fields.py::test_type_validation_int` - Validate integer field
- `tests/test_custom_fields.py::test_type_validation_date` - Validate date format
- `tests/test_custom_fields.py::test_enum_validation` - Validate enum values
- `tests/test_custom_fields.py::test_required_field` - Required field enforcement
- `tests/test_custom_fields.py::test_default_value` - Default value application
- `tests/test_custom_fields.py::test_export_with_custom_fields` - Export includes fields
- `tests/test_custom_fields.py::test_validate_all_fields` - Validate all requirements

## Technical Notes

### Custom Fields Configuration

```yaml
custom_fields:
  # Simple string field
  component:
    type: string
    description: "Component or module this requirement belongs to"
    required: false
    default: ""
    show_in_status: true

  # Enum field with allowed values
  risk_level:
    type: enum
    values: [low, medium, high, critical]
    description: "Risk assessment level"
    required: true
    default: medium

  # Integer field with validation
  complexity:
    type: int
    min: 1
    max: 10
    description: "Implementation complexity score"
    required: false

  # Date field
  target_date:
    type: date
    description: "Target completion date"
    required: false
    format: "%Y-%m-%d"  # ISO 8601 date

  # List field
  tags:
    type: list
    item_type: string
    description: "Requirement tags"
    required: false
    default: []

  # Boolean field
  customer_facing:
    type: bool
    description: "Is this requirement customer-facing?"
    required: false
    default: false
```

### Field Type Validators

```python
from datetime import datetime
from typing import Any

VALIDATORS = {
    "string": lambda v, f: isinstance(v, str),
    "int": lambda v, f: isinstance(v, int) and f.get("min", v) <= v <= f.get("max", v),
    "float": lambda v, f: isinstance(v, (int, float)),
    "bool": lambda v, f: isinstance(v, bool) or v in ["true", "false", "True", "False"],
    "date": lambda v, f: validate_date(v, f.get("format", "%Y-%m-%d")),
    "enum": lambda v, f: v in f.get("values", []),
    "list": lambda v, f: isinstance(v, list) or isinstance(v, str),
}

def validate_date(value: str, format: str) -> bool:
    try:
        datetime.strptime(value, format)
        return True
    except ValueError:
        return False
```

### Storage in RTM CSV

Custom fields are stored as additional columns in the RTM CSV:

```csv
req_id,status,...,component,risk_level,complexity,target_date,tags,customer_facing
REQ-PM-001,MISSING,...,sprint,medium,7,2024-02-15,"planning,core",true
REQ-PM-002,MISSING,...,analytics,low,5,2024-02-20,"reporting",false
```

### CLI Examples

```bash
$ rtmx field list
Name            Type    Required  Default     Description
component       string  no        ""          Component or module
risk_level      enum    yes       medium      Risk assessment level
complexity      int     no        -           Implementation complexity
target_date     date    no        -           Target completion date
tags            list    no        []          Requirement tags
customer_facing bool    no        false       Customer-facing?

$ rtmx field set REQ-PM-001 risk_level high
Set REQ-PM-001.risk_level = high

$ rtmx field set REQ-PM-001 complexity 15
Error: complexity must be between 1 and 10

$ rtmx field set REQ-PM-001 target_date 2024-02-15
Set REQ-PM-001.target_date = 2024-02-15

$ rtmx field get REQ-PM-001 risk_level
high

$ rtmx validate
Validating custom fields...
Error: REQ-PM-003 missing required field 'risk_level'
Error: REQ-PM-005 has invalid complexity value '12' (must be 1-10)
2 validation errors found
```

## Dependencies

None - this is an independent schema extension feature.

## Blocks

None - other features can optionally use custom fields.

# REQ-DX-003: Auto-Generate Requirement Specs

## Status: NOT_STARTED
## Priority: MEDIUM
## Phase: 4
## Effort: 0.75 weeks

## Description

Automatically generate requirement specification files from database entries using Jinja2 templates.

## Acceptance Criteria

- [ ] `rtmx setup --scaffold` creates spec files for all requirements
- [ ] Spec files use configurable Jinja2 template
- [ ] Default template includes: description, acceptance criteria, test cases
- [ ] Skip existing spec files (don't overwrite)
- [ ] `--force` flag to regenerate all specs
- [ ] Template customizable via `.rtmx/templates/requirement.md.j2`

## Default Template

```markdown
# {{ req_id }}: {{ requirement_text[:60] }}

## Status: {{ status }}
## Priority: {{ priority }}
## Phase: {{ phase }}

## Description

{{ requirement_text }}

## Acceptance Criteria

{% if target_value %}
- [ ] {{ target_value }}
{% else %}
- [ ] TBD
{% endif %}

## Test Cases

{% if test_module and test_function %}
- `{{ test_module }}::{{ test_function }}`
{% else %}
- No tests linked yet
{% endif %}

## Notes

{{ notes or "None" }}
```

## Usage

```bash
# Generate specs for all requirements without specs
rtmx setup --scaffold

# Force regenerate all specs (overwrites existing)
rtmx setup --scaffold --force

# Preview what would be created
rtmx setup --scaffold --dry-run
```

## Custom Template

Users can customize the template by creating:

```
.rtmx/templates/requirement.md.j2
```

Available template variables:
- `req_id`, `category`, `subcategory`
- `requirement_text`, `target_value`
- `test_module`, `test_function`
- `validation_method`, `status`
- `priority`, `phase`
- `notes`, `effort_weeks`
- `dependencies`, `blocks`
- `assignee`, `sprint`
- `started_date`, `completed_date`

## Files to Modify

- `src/rtmx/cli/setup.py` - Add --scaffold option
- NEW `src/rtmx/templates.py` - Template rendering

## Dependencies

- REQ-DX-001 (uses .rtmx/templates/)
- jinja2 already in optional dependencies (agents extra)

## Notes

This reduces manual work when setting up new projects or adding requirements.

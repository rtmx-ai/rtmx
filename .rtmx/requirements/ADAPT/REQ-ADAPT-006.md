# REQ-ADAPT-006: Monday.com Group-to-Category Mapping

## Metadata
- **Category**: ADAPT
- **Subcategory**: Monday
- **Priority**: MEDIUM
- **Phase**: 27
- **Status**: MISSING
- **Dependencies**: REQ-ADAPT-004, REQ-ADAPT-005
- **Blocks**: (none)

## Requirement

The Monday.com adapter shall support mapping Monday board groups to RTMX
categories, enabling structural alignment between the Monday board
organization and the RTMX requirements hierarchy.

## Rationale

Monday.com boards organize items into groups, which serve a similar
purpose to RTMX categories. Without group-to-category mapping, synced
items lose their organizational context in both directions.

## Design

### Configuration

```yaml
rtmx:
  adapters:
    monday:
      category_mapping:
        groups:
          "CLI": "CLI"
          "Server": "MCP"
          "Dashboard": "DASH"
          "Integrations": "ADAPT"
        default_category: "MISC"
```

### Behavior

- Pull: Monday group determines RTMX category
- Push: RTMX category determines target Monday group
- Unmapped groups use `default_category`
- `create_groups: true` creates missing Monday groups on push

## Acceptance Criteria

1. Monday groups map to RTMX categories per configuration.
2. RTMX categories map back to Monday groups for push sync.
3. Unmapped groups fall back to `default_category`.
4. `create_groups: true` creates missing groups via mutation.
5. Configuration validates mapped categories exist.

## Files to Create/Modify

- `internal/adapters/monday.go` -- Group mapping logic
- `internal/adapters/monday_test.go` -- Mapping tests

## Effort Estimate

0.5 weeks

## Test Strategy

- Table-driven mapping tests: group -> category, category -> group
- Default fallback: unmapped group uses default_category
- Group creation: verify mutation when create_groups enabled

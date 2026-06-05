# REQ-ADAPT-003: Asana Section-to-Category Mapping

## Metadata
- **Category**: ADAPT
- **Subcategory**: Asana
- **Priority**: MEDIUM
- **Phase**: 27
- **Status**: MISSING
- **Dependencies**: REQ-ADAPT-001, REQ-ADAPT-002
- **Blocks**: (none)

## Requirement

The Asana adapter shall support mapping Asana project sections to RTMX
categories and phases, enabling structural alignment between the Asana
project board and the RTMX requirements hierarchy.

## Rationale

Asana organizes tasks into sections within a project, similar to how RTMX
organizes requirements into categories and phases. Without structural
mapping, synced tasks lose their organizational context. Section-to-category
mapping preserves the hierarchical relationship during bidirectional sync.

## Design

### Configuration

```yaml
rtmx:
  adapters:
    asana:
      category_mapping:
        sections:
          "CLI Features": "CLI"
          "MCP Server": "MCP"
          "Dashboard": "DASH"
          "Integrations": "ADAPT"
        default_category: "MISC"
      phase_mapping:
        sections:
          "Sprint 1": 25
          "Sprint 2": 26
          "Backlog": 0
```

### Behavior

- When pulling from Asana: task section determines RTMX category
- When pushing to Asana: RTMX category determines target section
- Tasks in unmapped sections use `default_category`
- Requirements in unmapped categories go to a configurable default section

### Section CRUD

The adapter creates new sections in Asana if a required section does not
exist (when pushing a new category for the first time). This is controlled
by a `create_sections: true/false` config flag (default false).

## Acceptance Criteria

1. Asana sections map to RTMX categories per configuration.
2. RTMX categories map back to Asana sections for push sync.
3. Unmapped sections fall back to `default_category`.
4. `create_sections: true` creates missing Asana sections on push.
5. Phase mapping assigns correct phase number from section name.
6. Configuration validates that mapped categories exist in database.

## Files to Create/Modify

- `internal/adapters/asana.go` -- Section mapping logic
- `internal/adapters/asana_test.go` -- Mapping tests
- `internal/config/config.go` -- AsanaConfig mapping fields

## Effort Estimate

0.5 weeks

## Test Strategy

- Table-driven mapping tests: section -> category, category -> section
- Default fallback: unmapped section uses default_category
- Section creation: verify API call when create_sections enabled
- Validation: invalid category reference caught at config load

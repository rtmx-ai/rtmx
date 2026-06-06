# REQ-DASH-003: Requirement Detail Page with Edit

## Metadata
- **Category**: DASH
- **Subcategory**: View
- **Priority**: HIGH
- **Phase**: 28
- **Status**: MISSING
- **Dependencies**: REQ-DASH-001, REQ-DASH-002, REQ-API-002, REQ-API-003
- **Blocks**: (none)

## Requirement

The web dashboard shall provide a requirement detail page that displays
full requirement metadata, dependency visualization, and inline editing
of mutable fields (status, assignee, priority, sprint, notes).

## Rationale

Detail views with inline editing are the core interaction pattern in
project management tools. Users need to view complete requirement context
and make quick updates (assign to a sprint, change priority, update status)
without navigating to a separate edit form or using the CLI.

## Design

### Layout

```
+-- REQ-MCP-007: Response Size Logging -----------------------------------+
|                                                                          |
| Status: [MISSING v]    Priority: [P0 v]     Phase: 27                    |
| Assignee: [________]   Sprint: [________]   Effort: 0.5 weeks           |
| Started: --            Completed: --                                     |
|                                                                          |
| REQUIREMENT                                                              |
| RTMX MCP server shall log response byte count and estimated token count  |
| to stderr on every tools/call invocation, providing zero-cost            |
| observability into agent token consumption.                              |
|                                                                          |
| DEPENDENCIES                        BLOCKS                               |
| [COMPLETE] REQ-MCP-003 Read tools   [MISSING] REQ-MCP-008 Filtering     |
| [COMPLETE] REQ-MCP-006 Stdio        [MISSING] REQ-MCP-009 Size hints    |
|                                                                          |
| TEST INFO                                                                |
| Module: internal/adapters/mcp/server_test.go                             |
| Function: TestMCPResponseSizeLogging                                     |
| Method: Integration Test                                                 |
|                                                                          |
| [Save Changes]  [Cancel]                     [Back to List]              |
+--------------------------------------------------------------------------+
```

### Inline Editing

Mutable fields render as form controls:
- Status: dropdown select
- Priority: dropdown select
- Assignee: text input with autocomplete from known assignees
- Sprint: text input with version suggestion
- Notes: textarea

Changes are submitted via `PATCH /api/requirements/:id` (REQ-API-003).
htmx handles the form submission and swaps in the updated detail view.

### Dependency Links

Dependency and blocks lists are clickable links that navigate to the
linked requirement's detail page. Status badges are color-coded.

### Validation Feedback

Server-side validation errors (invalid status, blocked COMPLETE transition)
display as inline error messages near the affected field.

## Acceptance Criteria

1. Detail page displays all requirement metadata fields.
2. Status and priority render as dropdown selects with correct options.
3. Assignee and sprint render as editable text inputs.
4. Save button submits PATCH request and updates displayed values.
5. Validation errors display inline near the affected field.
6. Dependency and blocks links navigate to linked requirement detail.
7. Back to List button returns to the requirements table with filters preserved.
8. Cancel button reverts unsaved changes.
9. Unsaved changes trigger a confirmation prompt on navigation away.

## Files to Create/Modify

- `dashboard/detail.html` -- Detail page template
- `dashboard/components/edit-field.html` -- Inline edit component
- `internal/cmd/serve_api.go` -- PATCH handler validation error responses

## Effort Estimate

1 week

## Test Strategy

- Render test: all fields populated from API response
- Edit flow: change status, submit, verify PATCH request sent
- Validation: submit invalid status, verify inline error displayed
- Navigation: dependency links route to correct detail page
- Unsaved changes: navigate away, verify confirmation prompt

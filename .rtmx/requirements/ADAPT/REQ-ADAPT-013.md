# REQ-ADAPT-013: Unified Requirement ID Extraction

## Metadata
- **Category**: ADAPT
- **Subcategory**: Core
- **Priority**: HIGH
- **Phase**: 27
- **Status**: MISSING
- **Dependencies**: REQ-ADAPT-001, REQ-ADAPT-004, REQ-ADAPT-007
- **Blocks**: (none)

## Requirement

All adapters shall use a single, shared requirement ID extraction function
that correctly parses `REQ-<CATEGORY>-<NUMBER>` identifiers from any text
context: bracketed titles (`[REQ-CLI-001]`), `RTMX:` prefixed notes, and
inline mentions in descriptions. The category segment shall accept
uppercase alphanumeric characters (`[A-Z0-9]+`), not just letters.

## Rationale

Three independent extraction mechanisms exist across the adapter codebase,
each with different bugs:

1. **`extractReqID()` (Asana, Monday)**: String-searches for `"RTMX: "`
   prefix only. Silently returns empty string for `[REQ-CLI-001] Build CLI`
   titles -- the exact format these adapters use in `CreateItem()`. Monday
   tests never assert `RequirementID` so the bug is invisible.

2. **Regex in GitHub, Jira, GitLab `issueToItem()`**: Pattern
   `REQ-[A-Z]+-\d+` restricts the middle segment to pure letters. IDs
   like `REQ-E2E-010`, `REQ-V2-001`, or `REQ-K8S-001` are silently
   dropped. All current IDs happen to be pure letters, masking the bug.

3. **No shared function**: Each adapter reimplements extraction. A fix to
   one does not propagate to the others.

## Design

### Shared Extraction Function

Replace all three mechanisms with a single exported function in
`internal/adapters/adapter.go`:

```go
// reqIDRegex matches REQ-<CATEGORY>-<NUMBER> where CATEGORY is
// uppercase alphanumeric (e.g., CLI, MCP, E2E, V2, K8S).
var reqIDRegex = regexp.MustCompile(`REQ-[A-Z][A-Z0-9]*-\d+`)

// ExtractReqID finds the first RTMX requirement ID in text.
// It handles bracketed titles ([REQ-CLI-001]), RTMX: prefixes,
// and bare inline mentions.
func ExtractReqID(text string) string {
    if m := reqIDRegex.FindString(text); m != "" {
        return m
    }
    return ""
}
```

### Adapters Updated

| Adapter | Current | After |
|---------|---------|-------|
| Asana | `extractReqID()` (string search for "RTMX: ") | `ExtractReqID()` on notes, then title |
| Monday | `extractReqID()` on name only | `ExtractReqID()` on name |
| GitHub | inline regex `REQ-[A-Z]+-\d+` | `ExtractReqID()` |
| Jira | inline regex `REQ-[A-Z]+-\d+` | `ExtractReqID()` |
| GitLab | inline regex `REQ-[A-Z]+-\d+` | `ExtractReqID()` |

### Search Priority

For adapters with multiple text fields (Asana has notes + title, GitLab
has description + title), search the most specific field first (notes/
description), then fall back to title. This preserves the existing
convention where `RTMX: REQ-xxx` in notes takes precedence.

## Acceptance Criteria

1. Single `ExtractReqID()` function in `adapter.go` used by all adapters.
2. Pattern matches `REQ-[A-Z][A-Z0-9]*-\d+` (alphanumeric category).
3. Extracts from bracketed titles: `[REQ-CLI-001] Build CLI` -> `REQ-CLI-001`.
4. Extracts from RTMX prefix: `RTMX: REQ-MCP-001` -> `REQ-MCP-001`.
5. Extracts from inline text: `See REQ-E2E-010 for details` -> `REQ-E2E-010`.
6. Monday adapter `RequirementID` is no longer silently empty.
7. All adapters handle alphanumeric categories (E2E, V2, K8S).
8. Old `extractReqID()` function removed.
9. All existing adapter tests continue to pass.
10. New tests cover edge cases: multiple IDs (first wins), no ID, mixed case.

## Files to Create/Modify

- `internal/adapters/adapter.go` -- Add `ExtractReqID()` and `reqIDRegex`
- `internal/adapters/asana.go` -- Replace `extractReqID()` calls
- `internal/adapters/monday.go` -- Replace `extractReqID()` calls
- `internal/adapters/github.go` -- Replace inline regex
- `internal/adapters/jira.go` -- Replace inline regex
- `internal/adapters/gitlab.go` -- Replace inline regex
- `internal/adapters/adapter_test.go` -- Unit tests for `ExtractReqID()`

## Effort Estimate

0.25 weeks

## Test Strategy

- Table-driven unit tests for `ExtractReqID()` covering all input formats
- Verify Monday adapter `RequirementID` is now populated (regression)
- Verify alphanumeric IDs like `REQ-E2E-010` are extracted (regression)
- All existing adapter tests must continue to pass

# REQ-DASH-006: Release Planning View

## Metadata
- **Category**: DASH
- **Subcategory**: View
- **Priority**: HIGH
- **Phase**: 28
- **Status**: MISSING
- **Dependencies**: REQ-DASH-001, REQ-API-006, REQ-API-003
- **Blocks**: (none)

## Requirement

The web dashboard shall provide a release planning view that displays
version-scoped requirement summaries, gate status, and supports assigning
unversioned requirements to target versions via drag-and-drop or dropdown
selection.

## Rationale

Release planning -- deciding what goes into each version -- is a core PM
activity. The CLI `rtmx release scope` and `rtmx release assign` commands
exist but lack the visual overview needed for effective sprint planning.
A GUI view with drag-and-drop version assignment replaces the manual
`rtmx release assign` workflow.

## Design

### Layout

```
+-- Release Planning ---------------------------------------------------+
|                                                                        |
| +-- v1.0.0 (PASS) --------+  +-- v1.2.0 (FAIL) -------+  +-- Next --+|
| | 150/150 complete (100%)  |  | 47/50 complete (94%)   |  | Unversio||
| | Gate: PASS               |  | Gate: FAIL             |  | 41 reqs  ||
| | [View scope]             |  | [View scope]           |  |          ||
| +--------------------------+  |                        |  |          ||
|                               | Missing:               |  | REQ-TUI |
|                               |   REQ-MCP-007 (P0)     |  | REQ-DASH||
|                               |   REQ-MCP-008 (P0)     |  | REQ-ADAP||
|                               |   REQ-MCP-009 (HIGH)   |  | ...      ||
|                               +------------------------+  +----------+|
|                                                                        |
| ASSIGNMENT                                                             |
| Drag requirements from "Next" to a version, or select version from     |
| dropdown on requirement detail.                                        |
+------------------------------------------------------------------------+
```

### Version Cards

Each version shows:
- Version number with gate status badge (PASS/FAIL)
- Completion progress (X/Y, percentage, progress bar)
- List of incomplete requirements (if FAIL)
- Click to expand full requirement list

### Version Assignment

- Drag unversioned requirements from "Next" column to a version card
- This triggers `PATCH /api/requirements/:id` with `sprint: "v1.3.0"`
- Alternatively, use a dropdown on individual requirement cards

### Gate Check

Clicking "View scope" on a version card opens a modal showing the full
gate check output matching `rtmx release gate vX.Y.Z`.

## Acceptance Criteria

1. All versions display as cards with completion summary.
2. Gate status badge shows PASS (green) or FAIL (red).
3. FAIL versions list their incomplete requirements.
4. Unversioned requirements displayed in "Next" column.
5. Drag from Next to a version card assigns the version.
6. Dropdown assignment also works for individual requirements.
7. Version assignment persists via PATCH API.
8. "View scope" modal shows full gate check detail.
9. Progress bars reflect completion percentage.

## Files to Create/Modify

- `dashboard/releases.html` -- Release planning template
- `dashboard/js/releases.js` -- Version assignment logic
- `dashboard/components/version-card.html` -- Version summary card

## Effort Estimate

1 week

## Test Strategy

- Version cards: verify correct count and gate status per version
- Assignment: drag req to version, verify PATCH with sprint field
- Gate detail: verify modal content matches CLI gate output
- Unversioned: verify unassigned reqs appear in Next column
- Progress bar: verify width matches completion percentage

# REQ-PM-005: Epic and Initiative hierarchy

## Status: NOT STARTED
## Priority: MEDIUM
## Phase: 17
## Estimated Effort: 2.0 weeks

## Description

System shall support hierarchical requirement organization through Epic and Initiative types. This enables grouping related requirements under parent containers for portfolio-level planning and tracking. The hierarchy follows: Initiative > Epic > Requirement.

## Acceptance Criteria

- [ ] New `type` field added to RTM schema: `requirement` (default), `epic`, `initiative`
- [ ] New `parent` field links child items to parent epic/initiative
- [ ] `rtmx epic list` displays all epics with child requirement counts
- [ ] `rtmx epic show <epic_id>` displays epic details and child requirements
- [ ] `rtmx epic create <name> --description <desc>` creates new epic
- [ ] `rtmx epic link <req_id> <epic_id>` links requirement to epic
- [ ] `rtmx epic unlink <req_id>` removes requirement from epic
- [ ] `rtmx initiative list` displays all initiatives with epic counts
- [ ] `rtmx initiative show <init_id>` displays initiative with linked epics
- [ ] Epic status auto-calculated from child requirement statuses
- [ ] Initiative status auto-calculated from child epic statuses
- [ ] Progress percentage shown for epics/initiatives based on completion
- [ ] `rtmx tree` shows full hierarchy as ASCII tree
- [ ] Circular parent references are detected and rejected
- [ ] Orphan detection: requirements without parent epic (optional warning)

## Test Cases

- `tests/test_epic.py::test_epic_create` - Create new epic
- `tests/test_epic.py::test_epic_list` - List epics with counts
- `tests/test_epic.py::test_epic_show` - Show epic with children
- `tests/test_epic.py::test_epic_link_requirement` - Link requirement to epic
- `tests/test_epic.py::test_epic_unlink_requirement` - Unlink requirement from epic
- `tests/test_epic.py::test_epic_status_calculation` - Auto-calculate epic status
- `tests/test_epic.py::test_initiative_create` - Create initiative
- `tests/test_epic.py::test_initiative_link_epic` - Link epic to initiative
- `tests/test_epic.py::test_hierarchy_tree` - Display full hierarchy tree
- `tests/test_epic.py::test_circular_reference_detection` - Detect circular parents
- `tests/test_epic.py::test_progress_calculation` - Progress percentage accuracy

## Technical Notes

### Schema Additions

```csv
req_id,type,parent,category,...
INIT-001,initiative,,STRATEGIC,Strategic initiative for Q1
EPIC-001,epic,INIT-001,FEATURES,Feature epic for authentication
REQ-AUTH-001,requirement,EPIC-001,FEATURES,User login feature
REQ-AUTH-002,requirement,EPIC-001,FEATURES,Password reset feature
```

### ID Conventions

- Initiatives: `INIT-NNN`
- Epics: `EPIC-NNN` or category-prefixed `EPIC-AUTH-NNN`
- Requirements: `REQ-XXX-NNN` (unchanged)

### Status Calculation Rules

```python
def calculate_epic_status(epic_id: str) -> str:
    """Calculate epic status from child requirements."""
    children = get_children(epic_id)
    if not children:
        return "EMPTY"

    statuses = [c.status for c in children]
    if all(s == "COMPLETE" for s in statuses):
        return "COMPLETE"
    elif any(s in ["IN_PROGRESS", "PARTIAL"] for s in statuses):
        return "IN_PROGRESS"
    elif any(s != "MISSING" for s in statuses):
        return "PARTIAL"
    return "MISSING"

def calculate_progress(parent_id: str) -> float:
    """Calculate completion percentage."""
    children = get_children(parent_id)
    if not children:
        return 0.0
    complete = sum(1 for c in children if c.status == "COMPLETE")
    return (complete / len(children)) * 100
```

### CLI Output Examples

```bash
$ rtmx epic list

ID          Status       Progress  Children  Description
EPIC-001    IN_PROGRESS  60%       5         Authentication features
EPIC-002    COMPLETE     100%      3         API documentation
EPIC-003    MISSING      0%        8         Mobile support

$ rtmx tree

INIT-001: Q1 Strategic Goals [40%]
├── EPIC-001: Authentication [60%]
│   ├── REQ-AUTH-001: User login [COMPLETE]
│   ├── REQ-AUTH-002: Password reset [IN_PROGRESS]
│   └── REQ-AUTH-003: MFA support [MISSING]
├── EPIC-002: API Documentation [100%]
│   ├── REQ-DOC-001: OpenAPI spec [COMPLETE]
│   └── REQ-DOC-002: Developer guide [COMPLETE]
└── EPIC-003: Mobile Support [0%]
    └── (8 requirements)
```

### Configuration in rtmx.yaml

```yaml
hierarchy:
  types: [initiative, epic, requirement]
  warn_orphans: true  # Warn about requirements without parent
  auto_status: true   # Auto-calculate parent status
  id_patterns:
    initiative: "INIT-{number:03d}"
    epic: "EPIC-{number:03d}"
```

## Dependencies

None - this is an independent hierarchical enhancement.

## Blocks

- REQ-PM-006: Release management (releases can group epics/initiatives)

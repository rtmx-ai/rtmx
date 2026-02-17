# RTMX Database Migration Plan

## Executive Summary

This document defines the migration strategy for consolidating requirements from two repositories (rtmx and rtmx-go) into a single authoritative source in rtmx-go.

## Current State

| Repository | Database Location | Requirements | Purpose |
|------------|------------------|--------------|---------|
| rtmx (Python) | docs/rtm_database.csv | 95 | CLI + pytest plugin |
| rtmx-go | .rtmx/database.csv | 68 | Go CLI port |

## Target State

| Repository | Database Location | Requirements | Purpose |
|------------|------------------|--------------|---------|
| rtmx (Python) | .rtmx/database.csv | ~10 | Pytest plugin only |
| rtmx-go | .rtmx/database.csv | ~120 | Full CLI + all features |

---

## Phase 1: Remove Duplicates (Immediate)

### Action: Delete REQ-GO-072

REQ-GO-072 (Go native markers) duplicates REQ-LANG-003 from the Python repo.

**Before**: Two requirements describing the same thing
- REQ-LANG-003: "Go testing integration with helper functions and struct tags"
- REQ-GO-072: "Go CLI shall provide native Go test markers via rtmx.Req()"

**After**: Single requirement REQ-LANG-003 migrated to rtmx-go

### Action: Update REQ-GO-073 Dependencies

REQ-GO-073 (v0.1.0 release) currently depends on REQ-GO-072.
Change dependency to REQ-LANG-003 (which will migrate).

---

## Phase 2: Migrate Language Support (v0.1.0)

Migrate REQ-LANG-* requirements to rtmx-go database.

| Req ID | Status | Action |
|--------|--------|--------|
| REQ-LANG-001 | MISSING | Migrate to rtmx-go |
| REQ-LANG-002 | MISSING | Migrate to rtmx-go |
| REQ-LANG-003 | PARTIAL | Migrate to rtmx-go |
| REQ-LANG-004 | MISSING | Migrate to rtmx-go |
| REQ-LANG-005 | MISSING | Migrate to rtmx-go |
| REQ-LANG-006 | MISSING | Migrate to rtmx-go |
| REQ-LANG-007 | COMPLETE | Migrate to rtmx-go |

### Dependency Updates

After migration:
- REQ-GO-034 (markers command) should depend on REQ-LANG-007
- REQ-GO-073 (v0.1.0) should depend on REQ-LANG-003

---

## Phase 3: Migrate Collaboration & Security (v1.0.0)

| Category | Requirements | Phase |
|----------|--------------|-------|
| GIT_SYNC | REQ-GIT-001 through 006 | v1.0.0 |
| COLLABORATION | REQ-COLLAB-001 through 003 | v1.0.0 |
| ZERO_TRUST | REQ-ZT-001 through 003 | v1.0.0 |
| SECURITY | REQ-VERIFY-001 through 004 | v1.0.0 |

---

## Phase 4: Migrate Advanced Features (Post v1.0.0)

| Category | Requirements | Priority |
|----------|--------------|----------|
| PROJECT_MANAGEMENT | REQ-PM-001 through 008 | Medium |
| BDD | REQ-BDD-001 through 006 | Medium |
| CLAUDE_INTEGRATION | REQ-CLAUDE-001 through 004 | Low |

---

## Phase 5: Deprecate Python CLI Requirements

After v1.0.0, update Python repo:

1. **Mark as DEPRECATED**:
   - REQ-CLI-001, 002, 003
   - REQ-DX-001 through 006
   - REQ-UX-001 through 009
   - REQ-DIST-002

2. **Add superseded_by column** or note in each requirement

3. **Retain only pytest plugin requirements**:
   - REQ-PYTEST-001
   - REQ-BDD-INF-001 (if pytest-bdd stays in Python)

---

## Database Schema Alignment

Both databases currently use the same 21-column schema:

```
req_id,category,subcategory,requirement_text,target_value,test_module,test_function,
validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,
assignee,sprint,started_date,completed_date,requirement_file,external_id
```

Migration requires:
1. Copy rows from Python database to Go database
2. Update file paths in `test_module` and `requirement_file`
3. Verify no ID collisions (REQ-GO-* vs REQ-LANG-*, etc.)

---

## ID Namespace Strategy

### Current
- rtmx-go: REQ-GO-XXX
- rtmx: REQ-{category}-XXX (e.g., REQ-LANG-003, REQ-GIT-001)

### After Migration
Keep separate namespaces for clarity:
- REQ-GO-XXX: Go CLI core functionality
- REQ-LANG-XXX: Language support features
- REQ-GIT-XXX: Git integration features
- REQ-ZT-XXX: Zero-trust authentication
- REQ-PM-XXX: Project management
- etc.

This allows traceability back to original requirements.

---

## Website Roadmap Update

Update https://rtmx.ai/roadmap to reflect:

1. **Phase 13-14**: Go CLI v0.1.0 (current focus)
2. **Phase 14-15**: Multi-language support
3. **Phase 15-16**: Git native merge + CRDT sync
4. **Phase 17+**: Project management, BDD, Claude integration

---

## Timeline

| Milestone | Target | Actions |
|-----------|--------|---------|
| Week 1 | Now | Remove REQ-GO-072, migrate REQ-LANG-003 |
| Week 2-4 | v0.1.0 | Implement Go markers, first Go release |
| Week 5-8 | v0.2.0 | Migrate LANG-*, implement language parsers |
| Week 9-12 | v1.0.0 | Migrate GIT/COLLAB/ZT, full feature parity |
| Post v1.0 | | Deprecate Python CLI, migrate remaining |

---

## Verification

After each migration phase:

1. Run `rtmx status` in rtmx-go to verify database integrity
2. Run `rtmx health` to check for broken dependencies
3. Update both READMEs with current roadmap state
4. Update rtmx.ai/roadmap with progress

# REQ-GO-075: Cross-Repo Requirement Move and Clone

## Metadata
- **Category**: SYNC
- **Subcategory**: CrossRepo
- **Priority**: HIGH
- **Phase**: 17
- **Status**: MISSING
- **Dependencies**: REQ-GO-035
- **Blocks**: REQ-GO-076|REQ-GO-077

## Requirement

`rtmx move` shall transfer a requirement from the current project to a target rtmx-enabled repository, and `rtmx clone` shall fork a requirement into another project while preserving a bidirectional provenance link.

## Rationale

The RTMX ecosystem spans multiple repositories (rtmx, rtmx-sync, rtmx.ai, rtmx-go). When architectural decisions reclassify a requirement (e.g., a server-side repo discovers a feature belongs in the client-side repo), there is no automated way to transfer it. Today this requires manual CSV editing, spec file copying, and cross-repo PR creation — error-prone and unauditable.

This was discovered during the rtmx-sync audit when 8 requirements (WEB-005/006/008, RT-004/005/006, WEB-009, SITE-007) were reclassified as belonging to rtmx CLI or rtmx.ai but could only be marked with `external_id` pointers, not actually migrated.

## Design

### `rtmx move REQ-WEB-005 --to /path/to/rtmx`

1. Validate the requirement exists in source RTM
2. Validate the target directory is an rtmx-enabled project (has `.rtmx/` or `docs/rtm_database.csv`)
3. Read the requirement row from source CSV and spec file from source requirements dir
4. Map the requirement ID to the target project's ID scheme (e.g., `REQ-WEB-005` → `REQ-RTMX-NNN` or preserve original)
5. Add the requirement row to target CSV
6. Copy the spec `.md` file to target requirements dir
7. Set `external_id` in source to `{target_repo}/{new_id}` (provenance link)
8. Set `external_id` in target to `{source_repo}/{old_id}` (origin link)
9. Update source CSV status to indicate moved (or remove row with `--remove` flag)
10. Output summary of changes in both repos

### `rtmx clone REQ-WEB-005 --to /path/to/rtmx`

Same as move but:
- Source requirement stays in place (not removed)
- Source `external_id` updated to link to target
- Target `external_id` links back to source
- Both maintain independent status tracking

### Flags

- `--to PATH` — target repo path (required)
- `--id REQ-NEW-001` — override target requirement ID (optional, auto-generated if omitted)
- `--remove` — remove from source after move (default: keep as reference with external_id)
- `--branch NAME` — create a branch in target repo for PR workflow (see REQ-GO-076)
- `--dry-run` — preview changes without writing

## Acceptance Criteria

1. `rtmx move` transfers requirement row and spec file to target repo
2. `rtmx clone` creates a copy in target while preserving source
3. Bidirectional `external_id` links established in both repos
4. Target repo's ID scheme respected (auto-increment or explicit)
5. `--dry-run` shows changes without writing
6. Error if target is not an rtmx-enabled project
7. Error if requirement does not exist in source
8. Dependency references updated or flagged if they span repos

## Test Strategy

- **Test Module**: `internal/sync/crossrepo_test.go`
- **Test Function**: `TestMoveRequirement`, `TestCloneRequirement`
- **Validation Method**: Integration Test

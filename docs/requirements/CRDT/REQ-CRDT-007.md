# REQ-CRDT-007: Collaborative text fields shall use Y.Text for character-level merge

## Status: MISSING
## Priority: HIGH
## Phase: 10

## Description
Text fields (`requirement_text`, `notes`) shall use Y.Text CRDT type to enable character-level collaborative editing. This allows multiple users to type in the same field simultaneously without conflicts.

## Rationale
While Y.Map with LWW semantics (REQ-CRDT-006) handles concurrent edits to different fields, real-time collaboration requires character-level merge for text fields. Without Y.Text, if two users edit the same text field, only one user's changes survive. Y.Text enables Google Docs-style collaborative editing where both users' keystrokes merge seamlessly.

## Acceptance Criteria
- [ ] `requirement_text` field uses Y.Text instead of string
- [ ] `notes` field uses Y.Text instead of string
- [ ] Concurrent character insertions merge correctly
- [ ] Cursor positions preserved during merge
- [ ] Undo/redo works correctly with collaborative edits
- [ ] Text formatting (if supported) merges correctly

## Test Cases
- `tests/test_crdt.py::TestYText::test_concurrent_character_insert`
- `tests/test_crdt.py::TestYText::test_concurrent_text_edit_merge`
- `tests/test_crdt.py::TestYText::test_ytext_to_string_roundtrip`
- `tests/test_crdt.py::TestYText::test_cursor_position_preservation`

## Technical Notes
- pycrdt provides `Text` type for Y.Text operations
- Migration required: existing string fields → Y.Text
- Backward compatibility: read old string format, write Y.Text
- Performance consideration: Y.Text has higher overhead than string

## Migration Strategy
1. Add Y.Text support alongside existing string fields
2. On load: convert string to Y.Text if needed
3. On save: always write Y.Text format
4. Version flag in document metadata

## Dependencies
- REQ-CRDT-002: Y.Map structures
- REQ-CRDT-006: Basic concurrent merge (LWW)
- REQ-COLLAB-001: Sync server (for real-time collaboration)

## Blocks
- REQ-COLLAB-004: Conflict resolution UI

## Notes
This requirement was split from REQ-CRDT-006 because Y.Text only provides value when the sync server (Phase 10) enables real-time multi-user editing. The LWW merge in REQ-CRDT-006 is sufficient for offline and async workflows.

# REQ-DATA-001c: Authoring UX spike (encoding trade space)

## Metadata
- **Category**: DATA
- **Subcategory**: DOCUMENT_MODEL
- **Priority**: MEDIUM
- **Phase**: 34
- **Status**: COMPLETE
- **Dependencies**: REQ-DATA-001a
- **Blocks**: REQ-DATA-001f
- **External ID**:
- **ADR**: docs/adr/0007-requirement-document-model.md

## Requirement

Time-box a comparison of at least three authoring encodings for the same
fixture document model: (1) pure JSONL including rationale strings,
(2) structured records (JSONL candidate) plus Markdown narrative companion,
(3) Markdown with YAML frontmatter as canonical structured source. Score edit
friction, GitHub review, agent token cost, and diff noise. Selection criterion
is fitness for AC-to-test ATDD—not preference for a single file extension.

## Rationale

JSONL is a means, not the outcome. Authoring pain that blocks humans or agents
from maintaining ACs and bindings will defeat ATDD regardless of verify logic.

## Acceptance Criteria

1. Same fixture authored three ways; one-page scorecard with explicit criteria and winner/runner-up.
2. Demonstrate at least one agent-scale projection that returns metadata + ACs + bindings without bulk rationale.
3. Recommendation names primary authoring format for 2.0 and how AC IDs are edited day-to-day.
4. Scorecard states why the loser(s) fail ATDD maintainability if applicable.

## Test Strategy

Manual spike + checked-in scorecard artifact (Markdown under `docs/` or
`.rtmx/requirements/DATA/spikes/`). No production code required.

## Out of Scope

Implementing `rtmx spec edit`; choosing sync encoding; flipping defaults.

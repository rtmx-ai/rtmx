# REQ-DATA-001f decision gate — Accept ADR-0007

**Date:** 2026-09-19  
**Outcome:** **GO** — accept the Requirement Document Model path to AC-to-test ATDD.  
**Selection criterion:** fitness for AC-to-test ATDD and per-AC verify gaps — **not** uniformity of a single file extension.

## Spike gate answers (ADR-0007)

| Question | Evidence | Result |
|---|---|---|
| Schema | `docs/schemas/requirement-document-v0.schema.json` + fixtures; `AC_ID_STABILITY.md` | **Pass** (001a) |
| Verify | `internal/docmodel/acverify` + `ATDD_POLICY_v0.md` (`atdd-required-all-v0`) | **Pass** (001b) |
| Authoring | Scorecard winner: JSONL structured + MD companion | **Pass** (001c) |
| Migration | Round-trip spike + `MIGRATE_ROUNDTRIP_LIMITS.md` | **Pass** (001d) |
| Sync | Additive `1.1-ac` sketch; no silent AC drop plan | **Pass** (001e) |
| MCP/token | `projection-agent-atdd-v0.json` + MCP sketch in SYNC_MCP_IMPACT_SPIKE | **Pass** (001e/001c) |

## Decisions

1. **Canonical structured encoding (2.0):** JSONL (or equivalent NDJSON record stream) for identity, `acs[]`, and `test_bindings[]`.
2. **Prose strategy:** optional Markdown companion via `narrative_path` / `requirement_file`; never required on the verify/MCP hot path.
3. **Completeness policy (default for AC-aware verify):** `atdd-required-all-v0` — all **required** ACs must have passing evidence before COMPLETE; optional ACs do not block.
4. **v1 compatibility:** CSV remains default on-disk format until 2.0.0; `rtmx migrate` (and CSV export projection) are mandatory before any default flip.
5. **Sync:** no breaking sync-protocol-v1 change for managed-sync launch; nested ACs land later as an **additive** extension (`docs/schemas/sync-protocol-v1.1-ac-extension.stub.json`).
6. **Encoding criterion statement:** JSONL+MD was chosen because it maximizes AC/binding editability and agent projection hygiene for ATDD; pure JSONL-with-rationale and MD-frontmatter-only lost on token cost and ID-drift risk (001c scorecard).

## Implementation backlog (go)

| ID | Scope |
|---|---|
| REQ-DATA-002 | Opt-in `rtmx verify` AC-matrix path using document model + `atdd-required-all-v0` |
| REQ-DATA-003 | `rtmx migrate` CSV+MD → structured JSONL+companion; idempotent `ac_id` |
| REQ-DATA-004 | Read/write helpers and CLI projection for structured store (no default flip) |
| REQ-DATA-005 | Implement additive sync `1.1-ac` maps behind capability flag (post managed-sync) |
| REQ-DATA-006 | MCP tool payloads: AC gaps + bindings; omit rationale by default |

## Explicitly deferred (remains v1)

- Monorepo-wide migration of all projects.
- Making COMPLETE globally mean “all ACs green” in default `rtmx verify` before DATA-002 ships.
- Replacing sync-protocol-v1 on production managed Sync.
- Dropping Markdown GitHub rendering.

## Parent closure

REQ-DATA-001 closes with this gate: ADR Accepted, schema published, spike answers recorded, no production default flip in this change, managed-sync unblocked.

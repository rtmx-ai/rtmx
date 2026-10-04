# Sync + MCP impact spike (REQ-DATA-001e)

Sketch only. No production rtmx-sync schema migration and no Yjs nested-type
cutover in this change.

## 1. Gap analysis vs sync-protocol-v1 `requirement_fields`

Source: `system/contracts/sync-protocol-v1.json` →
`crdt_document_schema.requirement_fields`.

| Document-model field (REQ-DATA-001a) | On v1 wire today | Gap |
|---|---|---|
| `req_id`, `category`, `subcategory`, `requirement_text`, `status`, `priority`, `phase`, `assignee`, `external_id` | Present (scalars / Y.Text) | None for identity spine |
| `dependencies`, `blocks`, `effort_weeks`, `target_value`, `requirement_file` | **Absent** | Soft gap; often recoverable from CSV import, not live-collaborative |
| `acs[]` (`ac_id`, `statement`, `required`, `severity`) | **Absent** | **Hard gap** — nested ACs silently dropped on sync |
| `test_bindings[]` | **Absent** | **Hard gap** — AC↔test links not collaborative |
| Per-AC evidence / gap status | **Absent** | Derived locally today; not shared |
| Narrative / rationale | Partially via `notes` Y.Text | Coarse; not separated from verify surface |
| `schema_version` on requirement | Only on room `metadata` | Room-level only; no per-req document version |

**Conclusion:** v1 can sync req-level status and collaborative prose (`requirement_text`, `notes`) but **cannot** represent AC-first ATDD without an additive extension. Managed-sync launch continues on v1 scalars; nested ACs remain local/CSV+MD until a later protocol bump.

## 2. Proposal: what to sync (subset) + tradeoffs

### Chosen subset for a future additive extension (`schema_version: "rtm-doc/v0"`)

Sync **structured, small** fields; keep narrative out of the hot map:

| Field | Sync? | Rationale |
|---|---|---|
| `acs[]` identity + `statement` + `required` | **Yes** | Required for shared AC gaps and conflict-aware edits |
| Per-AC `evidence` / last-known pass\|fail\|gap | **Optional / derived** | Prefer recompute from CI results projection over CRDT write storms |
| `test_bindings[]` | **Yes (compact)** | `binding_id`, `ac_id`, `test_module`, `test_function` or marker — no fixtures |
| Full rationale / MD body | **No** | Token + bandwidth; use `narrative_path` or out-of-band git |
| Req-level `status` | **Yes (existing)** | Keep v1 field; aggregation policy stays local until clients agree |

### Encoding on the CRDT (sketch)

Prefer **additive**, not a breaking rename of v1 maps:

1. Keep `requirements: Y.Map<req_id, RequirementV1Scalars>`.
2. Add optional sibling maps keyed by `req_id`:
   - `requirement_acs: Y.Map<req_id, Y.Array|Y.Map of AC records>`
   - `requirement_bindings: Y.Map<req_id, Y.Array of binding records>`
3. Bump room `metadata.schema_version` to advertise support (e.g. `1.1.0-ac`
   or document `rtm-doc/v0` capability flag). Old clients ignore unknown maps.

### Bandwidth

- AC statements are short; typical req has 2–5 ACs → low KB per req.
- Bindings are identifiers only; avoid embedding test bodies.
- Omitting rationale keeps room updates well under `max_message_size_bytes`.

### Conflicts

- **AC statement edits:** last-writer-wins on AC record fields, or Y.Text per
  `statement` if collaborative editing of AC prose is required later.
- **`ac_id` stability:** treat `ac_id` as immutable key; renames = delete+add
  with migrate tool, not in-place ID mutation (aligns with AC_ID_STABILITY).
- **Bindings:** set-union by `binding_id`; duplicate ac_id bindings are a
  verify concern, not a CRDT merge concern.
- **Status vs AC evidence:** do not CRDT-merge derived COMPLETE; clients apply
  `atdd-required-all-v0` (or successor) locally from evidence + required flags.

### Alternatives rejected for v1.1 sketch

- **Sync full Markdown bodies in Y.Text per req:** high conflict noise; fails
  MCP token goals (see §3).
- **Sync only AC status bits without statements:** agents cannot author or
  understand gaps without text; ATDD incomplete.
- **Break v1 by replacing Requirement shape:** unnecessary for launch; see §4.

Optional stub: `docs/schemas/sync-protocol-v1.1-ac-extension.stub.json`.

## 3. MCP projection sketch (AC gaps, no bulk rationale)

Align with authoring spike projection
(`.rtmx/requirements/DATA/spikes/authoring-ux/projection-agent-atdd-v0.json`).

### `rtmx_backlog` / `next`-style item (additive fields)

```json
{
  "req_id": "REQ-AUTH-001",
  "status": "PARTIAL",
  "priority": "HIGH",
  "requirement_text": "Users shall authenticate before accessing protected rooms.",
  "acs": [
    {"ac_id": "AC-1", "statement": "…", "required": true, "evidence": "pass"},
    {"ac_id": "AC-2", "statement": "…", "required": true, "evidence": "gap"}
  ],
  "ac_gap_count": 1,
  "test_bindings": [
    {"binding_id": "TB-1", "ac_id": "AC-1", "test_function": "TestRoomRejectsMissingToken"}
  ]
}
```

### `rtmx_verify` / markers-style summary

```json
{
  "req_id": "REQ-AUTH-001",
  "aggregated_status": "PARTIAL",
  "policy": "atdd-required-all-v0",
  "gaps": [{"ac_id": "AC-2", "reason": "no_passing_binding"}],
  "omit": ["rationale", "notes", "requirement_file_body"]
}
```

**Rule:** default MCP tool payloads MUST NOT include companion Markdown body or
embedded `rationale` strings. Opt-in flag (future) for narrative fetch.

## 4. Managed-sync launch constraint

**No breaking sync-protocol-v1 change is required for managed-sync launch.**

- Production rooms continue with existing `requirement_fields` scalars.
- Entitlement, Checkout, and CLI write paths do not depend on nested ACs.
- Document-model spikes (001a–001e) and any future `1.1-ac` extension are
  **orthogonal** to monetization / managed Sync go-live.
- CSV remains the v1 on-disk default until REQ-DATA-001f accepts 2.0 defaults.

## 5. Explicit non-implementation (this spike)

- No rtmx-sync production migration.
- No Yjs nested type deployed to Cloud Run.
- Stub JSON is documentation-only (not CI-enforced contract yet).

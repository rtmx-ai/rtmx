# Authoring UX scorecard (REQ-DATA-001c)

Same fixture: `REQ-AUTH-001` with two ACs, two bindings, and a rationale
paragraph. Encodings under this directory:

| Encoding | Path |
|---|---|
| (1) Pure JSONL (rationale embedded) | `jsonl-pure/REQ-AUTH-001.jsonl` |
| (2) JSONL + Markdown companion | `jsonl-plus-md/` |
| (3) Markdown + YAML frontmatter canonical | `md-frontmatter/REQ-AUTH-001.md` |

Agent-scale projection (metadata + ACs + bindings, no rationale):
`projection-agent-atdd-v0.json`.

## Criteria (1–5; higher is better for ATDD maintainability)

| Criterion | Weight | (1) Pure JSONL | (2) JSONL + MD | (3) MD frontmatter |
|---|---|---|---|---|
| Edit friction (add/rename AC, bind test) | 3 | 2 — edit long single-line JSON; easy to break | **5** — structured record for ACs; narrative isolated | 4 — YAML list edits are clear; frontmatter parsers vary |
| GitHub review readability | 2 | 1 — opaque JSONL diffs | **5** — MD narrative reviews well; JSONL hunks small | 4 — rendered body good; frontmatter diffs noisy |
| Agent token cost (load verify surface) | 3 | 2 — rationale always in the structured blob | **5** — project away `narrative_path` / MD body | 3 — must strip body after frontmatter parse |
| Diff noise (status/AC churn) | 2 | 2 — whole-line rewrites common | **4** — AC changes stay in JSONL; rationale MD separate | 3 — YAML indentation churn on small edits |
| AC-ID day-to-day edit story | 2 | 3 — `acs[].ac_id` in JSON | **5** — same; MD never owns IDs | 4 — IDs in frontmatter; risk of body list drift |

**Weighted totals:** (1) **20** · (2) **48** · (3) **36** (max 60).

## Winner / runner-up

- **Winner: (2) JSONL structured records + Markdown narrative companion.**
- **Runner-up: (3) Markdown + YAML frontmatter.**

## Recommendation for rtmx 2.0 primary authoring

**Primary authoring format:** structured JSONL (or equivalent record stream)
holding identity, `acs[]`, and `test_bindings[]`, with optional
`narrative_path` Markdown companion for rationale and history.

**Day-to-day AC ID editing:** edit `acs[].ac_id` and `statement` only in the
structured record. Never mint IDs in Markdown prose lists. Ordinal scheme
`AC-1`… per requirement (see `docs/schemas/AC_ID_STABILITY.md`); rematerialize
only when statement text changes under that policy.

**Agent verify default projection:** load structured fields only (as in
`projection-agent-atdd-v0.json`); omit rationale / companion body unless an
explicit `--include-narrative` path is requested.

## Why losers fail ATDD maintainability

- **(1) Pure JSONL with embedded rationale:** every agent/verify load pays for
  narrative tokens; GitHub review of AC/binding edits is hostile; one bad comma
  corrupts ACs and rationale together. Fails the “AC-to-test ATDD without bulk
  prose” selection criterion even though the schema is fine.
- **(3) MD+frontmatter as sole canonical store:** workable for humans, but
  agents must always parse and strip body; frontmatter indentation diffs
  obscure AC edits; dual lists (YAML `acs` vs accidental Markdown checkboxes)
  invite ID drift. Acceptable as an *import/export* view, not as the sole
  source of truth for bindings.

## Out of scope (per requirement)

No `rtmx spec edit` implementation; no sync encoding choice; no default flip.

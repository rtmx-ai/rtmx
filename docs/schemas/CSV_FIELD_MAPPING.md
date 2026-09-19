# CSV column → Requirement Document Model mapping (REQ-DATA-001a AC3)

Spike mapping for identity, status, dependencies, and related scalars.
On-disk encoding is TBD; this table is the logical field bridge only.

| CSV column (`database.csv`) | Schema field | Notes |
|---|---|---|
| `req_id` | `requirements[].req_id` | Required; identity key |
| `category` | `requirements[].category` | Optional scalar |
| `subcategory` | `requirements[].subcategory` | Optional scalar |
| `requirement_text` | `requirements[].requirement_text` | Required statement |
| `target_value` | `requirements[].target_value` | Optional |
| `status` | `requirements[].status` | Enum aligned with CSV statuses |
| `priority` | `requirements[].priority` | Free-form string (HIGH/P0/…) |
| `phase` | `requirements[].phase` | Integer or string |
| `effort_weeks` | `requirements[].effort_weeks` | Number |
| `dependencies` | `requirements[].dependencies[]` | Pipe-split in CSV → string array |
| `blocks` | `requirements[].blocks[]` | Pipe-split in CSV → string array |
| `external_id` | `requirements[].external_id` | Cross-project / monorepo link |
| `requirement_file` | `requirements[].requirement_file` | Narrative companion path |
| `test_module` | `requirements[].test_bindings[].test_module` | Primary CSV binding projects to req-level or first AC |
| `test_function` | `requirements[].test_bindings[].test_function` | Same |
| `notes` | `requirements[].test_bindings[].notes` or narrative | Spike: prefer binding notes for verify hints |
| `assignee` / `sprint` / dates | *(deferred)* | Not required for AC-to-test ATDD identity |

## Nested fields with no CSV column today

| Schema field | Source today | Notes |
|---|---|---|
| `acs[].ac_id` | Derived from Markdown AC lists | Ordinal scheme `AC-N` (see AC4) |
| `acs[].statement` | Markdown acceptance criteria | |
| `test_bindings[]` | CSV `test_*` + markers | Multiple bindings per req/AC |

## Out of production scope (REQ-DATA-001a AC5)

This mapping does **not** authorize rewriting `internal/database` or changing
CLI defaults. It documents the spike contract only.

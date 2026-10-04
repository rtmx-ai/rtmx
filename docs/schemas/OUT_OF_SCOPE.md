# REQ-DATA-001a out of scope (AC5)

Closing this spike does **not**:

- Rewrite `internal/database` CSV production read/write paths
- Change `rtmx` CLI defaults or verify promotion rules
- Choose on-disk encoding (JSONL vs Markdown+frontmatter vs other)
- Modify sync-protocol-v1

Artifacts live under `docs/schemas/` and `internal/docmodel` (validate-only).

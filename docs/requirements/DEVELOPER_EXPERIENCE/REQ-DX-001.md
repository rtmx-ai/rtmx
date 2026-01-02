# REQ-DX-001: Implement .rtmx/ Directory Structure

## Status: NOT_STARTED
## Priority: HIGH
## Phase: 4
## Effort: 1.5 weeks

## Description

Consolidate all rtmx artifacts into a `.rtmx/` directory, following the pattern of `.venv`, `.claude`, and `.git`.

## Acceptance Criteria

- [ ] New projects created with `rtmx init` use `.rtmx/` structure
- [ ] Config discovery checks `.rtmx/config.yaml` first
- [ ] Database discovery checks `.rtmx/database.csv` first
- [ ] Requirements stored in `.rtmx/requirements/`
- [ ] Cache directory `.rtmx/cache/` for generated artifacts
- [ ] `.rtmx/.gitignore` auto-created to ignore cache/
- [ ] All existing commands work with new structure

## Directory Structure

```
project/
├── .rtmx/                          # RTMX home directory
│   ├── config.yaml                 # Configuration (was rtmx.yaml)
│   ├── database.csv                # RTM database (was docs/rtm_database.csv)
│   ├── requirements/               # Requirement specs (was docs/requirements/)
│   │   └── CATEGORY/
│   │       └── REQ-XX-NNN.md
│   ├── cache/                      # Generated artifacts (NEW)
│   │   ├── api-docs/               # pdoc output
│   │   ├── reports/                # Health reports, diffs
│   │   └── snapshots/              # Point-in-time exports
│   ├── templates/                  # User templates (NEW)
│   │   └── requirement.md.j2
│   └── .gitignore                  # Ignore cache/
```

## Benefits

1. **Single directory** - Easy to identify, backup, or remove
2. **Clear ownership** - All rtmx artifacts in one place
3. **Gitignore-friendly** - One pattern to manage
4. **Portable** - Copy `.rtmx/` to replicate setup
5. **IDE-friendly** - Can collapse entire directory

## Files to Modify

- `src/rtmx/config.py` - Config discovery order
- `src/rtmx/cli/init.py` - Create new structure
- `src/rtmx/cli/setup.py` - Setup with new paths
- `src/rtmx/models.py` - Database path resolution

## Default .gitignore Content

```gitignore
# Ignore cache directory
cache/

# Keep everything else
!.gitignore
```

## Config Discovery Order

1. `.rtmx/config.yaml` (new)
2. `rtmx.yaml` (legacy, with deprecation warning)
3. Parent directory search (both patterns)

## Blocks

- REQ-DX-002 (migration depends on structure)
- REQ-DX-003 (scaffold depends on structure)
- REQ-DX-004 (docs depend on cache/)

## Notes

This is the foundational change that other DX improvements depend on.

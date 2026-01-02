# RTMX Lifecycle Specification

This document specifies the complete lifecycle of RTMX within a project repository, including initialization, configuration, operations, and removal.

## Overview

RTMX manages requirements traceability through a file-based system designed for Git version control. The lifecycle consists of five phases:

1. **Initialization** - Set up RTM structure in a project
2. **Configuration** - Customize behavior via `rtmx.yaml`
3. **Operations** - Day-to-day requirement management
4. **Integration** - Connect with external systems and AI agents
5. **Removal** - Clean uninstallation from a project

## Project Structure

After initialization, RTMX creates the following structure:

```
project/
├── .rtmx/                        # RTMX state directory (optional)
│   ├── templates/                # Agent prompt templates
│   ├── cache/                    # Local cache (gitignored)
│   └── backups/                  # Auto-backups before destructive ops
├── docs/
│   ├── rtm_database.csv          # Requirements database (source of truth)
│   └── requirements/             # Requirement specification files
│       └── {CATEGORY}/
│           └── {REQ-ID}.md
└── rtmx.yaml                     # Configuration file
```

### File Purposes

| File/Directory | Purpose | Version Control |
|----------------|---------|-----------------|
| `rtmx.yaml` | Project configuration | Yes |
| `docs/rtm_database.csv` | Requirements database | Yes |
| `docs/requirements/` | Detailed specifications | Yes |
| `.rtmx/templates/` | Agent prompt templates | Yes |
| `.rtmx/cache/` | Temporary state | No (.gitignore) |
| `.rtmx/backups/` | Operation backups | Optional |

---

## Phase 1: Initialization

### Command: `rtmx init`

Creates the RTM structure in the current project.

```bash
rtmx init [--force] [--schema {core|phoenix}]
```

#### Options

| Option | Description |
|--------|-------------|
| `--force` | Overwrite existing files |
| `--schema` | Schema type: `core` (20 cols) or `phoenix` (extended) |

#### Behavior

1. **Check for existing files**
   - If `rtmx.yaml` exists and `--force` not set → Error
   - If `docs/rtm_database.csv` exists and `--force` not set → Error

2. **Create directory structure**
   ```
   docs/
   docs/requirements/
   docs/requirements/EXAMPLE/
   .rtmx/
   .rtmx/templates/
   .rtmx/cache/
   .rtmx/backups/
   ```

3. **Create configuration file** (`rtmx.yaml`)
   - Default settings for database path, requirements directory
   - Pytest marker configuration
   - Disabled adapters (GitHub, Jira)

4. **Create initial database** (`docs/rtm_database.csv`)
   - Header row with schema columns
   - Sample requirement `REQ-EX-001`

5. **Create sample specification** (`docs/requirements/EXAMPLE/REQ-EX-001.md`)
   - Template showing expected format

6. **Update .gitignore**
   - Add `.rtmx/cache/` if not present

#### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | File exists (use --force) |
| 2 | Permission denied |
| 3 | Invalid schema type |

#### Example Output

```
$ rtmx init
Created rtmx.yaml
Created docs/rtm_database.csv (1 sample requirement)
Created docs/requirements/EXAMPLE/REQ-EX-001.md

Next steps:
  1. Edit rtmx.yaml to customize settings
  2. Run 'rtmx status' to see current state
  3. Add requirements to docs/rtm_database.csv
```

---

## Phase 2: Configuration

### File: `rtmx.yaml`

The configuration file controls all RTMX behavior.

```yaml
rtmx:
  # Core settings
  database: docs/rtm_database.csv
  requirements_dir: docs/requirements
  schema: core  # core | phoenix

  # Pytest integration
  pytest:
    marker_prefix: "req"
    register_markers: true

  # AI agent integration
  agents:
    claude:
      enabled: true
      config_path: "CLAUDE.md"
    cursor:
      enabled: true
      config_path: ".cursorrules"
    copilot:
      enabled: true
      config_path: ".github/copilot-instructions.md"
    template_dir: ".rtmx/templates/"

  # External service adapters
  adapters:
    github:
      enabled: false
      repo: "org/repo"
      token_env: "GITHUB_TOKEN"
      labels: ["requirement"]
      status_mapping:
        open: MISSING
        closed: COMPLETE

    jira:
      enabled: false
      server: "https://company.atlassian.net"
      project: "PROJ"
      token_env: "JIRA_API_TOKEN"
      email_env: "JIRA_EMAIL"
      status_mapping:
        "To Do": MISSING
        "In Progress": PARTIAL
        "Done": COMPLETE

  # Sync behavior
  sync:
    conflict_resolution: manual  # manual | prefer-local | prefer-remote
    auto_backup: true

  # MCP server
  mcp:
    enabled: false
    port: 3000
    host: localhost
```

### Configuration Discovery

RTMX searches for configuration in this order:

1. `--config` flag (explicit path)
2. `rtmx.yaml` in current directory
3. `rtmx.yml` in current directory
4. Search parent directories up to git root
5. Default values if no config found

### Validation

Configuration is validated on load:

- Required fields must be present
- Paths must be valid
- Schema must be recognized
- Status mappings must use valid Status values

---

## Phase 3: Operations

### 3.1 Status and Reporting

#### `rtmx status`

Show completion status summary.

```bash
rtmx status [-v|-vv|-vvv] [--json OUTPUT]
```

| Verbosity | Shows |
|-----------|-------|
| (none) | Overall completion percentage |
| `-v` | Category breakdown |
| `-vv` | Subcategory breakdown |
| `-vvv` | All requirements |

#### `rtmx backlog`

Show incomplete requirements prioritized by criticality.

```bash
rtmx backlog [--phase N] [--view {all|critical|quick-wins|blockers}] [--limit N]
```

| Option | Description |
|--------|-------------|
| `--phase N` | Filter by phase number |
| `--view` | View mode: all (default), critical (path), quick-wins, blockers |
| `--limit N` | Limit items shown (default: 10) |

### 3.2 Requirement Management

#### Adding Requirements

1. **Manual**: Edit `docs/rtm_database.csv` directly
2. **From tests**: `rtmx from-tests --update`
3. **From external**: `rtmx sync github --import`
4. **Bootstrap**: `rtmx bootstrap --from-tests`

#### Updating Requirements

```python
from rtmx import RTMDatabase, Status

db = RTMDatabase.load("docs/rtm_database.csv")
db.update("REQ-SW-001", status=Status.COMPLETE)
db.save()
```

#### Removing Requirements

```bash
rtmx remove REQ-SW-001 [--force] [--no-backup]
```

Behavior:
1. Create backup in `.rtmx/backups/`
2. Remove from `docs/rtm_database.csv`
3. Optionally remove specification file

### 3.3 Dependency Management

#### `rtmx deps`

Show dependency graph.

```bash
rtmx deps [--req REQ-ID] [--category CAT] [--format {tree|dot|json}]
```

#### `rtmx cycles`

Detect circular dependencies.

```bash
rtmx cycles [--verbose]
```

Exit codes:
- 0: No cycles
- 1: Cycles detected

#### `rtmx reconcile`

Check and fix dependency reciprocity.

```bash
rtmx reconcile [--execute] [--dry-run]
```

Ensures: If A blocks B, then B depends on A.

### 3.4 Test Integration

#### `rtmx from-tests`

Discover requirements from pytest markers.

```bash
rtmx from-tests [PATH] [--all] [--missing] [--update]
```

| Option | Description |
|--------|-------------|
| `--all` | Show all discovered markers |
| `--missing` | Show markers not in RTM |
| `--update` | Update RTM with test info |

#### Pytest Markers

```python
@pytest.mark.req("REQ-SW-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_feature():
    pass
```

### 3.5 Validation

#### `rtmx health`

Run integration health checks on RTM database.

```bash
rtmx health [--format {terminal|json|ci}] [--strict] [--check NAME]
```

| Option | Description |
|--------|-------------|
| `--format` | Output format: terminal (default), json, ci |
| `--strict` | Treat warnings as errors |
| `--check` | Run specific checks only (can repeat) |

Checks performed:
- Required fields present
- Valid status/priority values
- No duplicate IDs
- Dependencies reference existing requirements
- Reciprocity consistency
- Test coverage gaps
- Phase ordering

Exit codes:
- 0: Healthy
- 1: Warnings (with --strict)
- 2: Errors

---

## Phase 4: Integration

### 4.1 AI Agent Integration

#### `rtmx install`

Install RTMX prompts into AI agent configurations.

```bash
rtmx install [--agents {claude,cursor,copilot}] [--all] [--dry-run] [--force]
```

Injected content:
- Quick command reference
- Test marker documentation
- Workflow guidance

#### `rtmx uninstall`

Remove RTMX prompts from agent configurations.

```bash
rtmx uninstall [--agents {claude,cursor,copilot}] [--all]
```

### 4.2 External Service Sync

#### `rtmx sync`

Synchronize with GitHub Issues or Jira.

```bash
rtmx sync {github|jira} [--import] [--export] [--bidirectional] [--dry-run]
```

| Mode | Description |
|------|-------------|
| `--import` | Pull from external → RTM |
| `--export` | Push RTM → external |
| `--bidirectional` | Two-way sync |

Conflict resolution modes:
- `prefer-local`: RTM wins
- `prefer-remote`: External wins
- `manual`: Report for user decision

### 4.3 Bootstrap

#### `rtmx bootstrap`

Generate RTM from existing project artifacts.

```bash
rtmx bootstrap [--from-tests] [--from-github] [--from-jira] [--merge] [--dry-run]
```

### 4.4 MCP Server

#### `rtmx mcp-server`

Start Model Context Protocol server for AI integration.

```bash
rtmx mcp-server [--port 3000] [--host localhost] [--daemon]
```

---

## Phase 5: Removal

### Command: `rtmx uninstall`

Remove RTMX from a project.

```bash
rtmx uninstall [--all] [--keep-data] [--force]
```

#### Options

| Option | Description |
|--------|-------------|
| `--all` | Remove all RTMX files (config + data) |
| `--keep-data` | Keep rtm_database.csv and requirements/ |
| `--force` | Don't prompt for confirmation |

#### Behavior

1. **Backup current state** (unless `--force`)
   - Copy to `.rtmx/backups/uninstall-{timestamp}/`

2. **Remove agent integrations**
   - Remove RTMX sections from CLAUDE.md, .cursorrules, etc.

3. **Remove configuration**
   - Delete `rtmx.yaml`

4. **Remove state directory**
   - Delete `.rtmx/` entirely

5. **Optionally remove data** (if `--all` and not `--keep-data`)
   - Delete `docs/rtm_database.csv`
   - Delete `docs/requirements/`

#### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Cancelled by user |
| 2 | Backup failed |

---

## Data Model

### Requirement Schema (Core - 20 columns)

| Column | Type | Required | Description |
|--------|------|----------|-------------|
| `req_id` | string | Yes | Unique identifier (REQ-XX-NNN) |
| `category` | string | Yes | High-level grouping |
| `subcategory` | string | No | Detailed classification |
| `requirement_text` | string | Yes | Human-readable description |
| `target_value` | string | No | Quantitative criteria |
| `test_module` | string | No | Test file path |
| `test_function` | string | No | Test function name |
| `validation_method` | string | No | Testing approach |
| `status` | enum | Yes | COMPLETE/PARTIAL/MISSING/NOT_STARTED |
| `priority` | enum | No | P0/HIGH/MEDIUM/LOW |
| `phase` | integer | No | Development phase (≥1) |
| `notes` | string | No | Additional context |
| `effort_weeks` | float | No | Estimated effort |
| `dependencies` | list | No | Pipe-separated req IDs |
| `blocks` | list | No | Pipe-separated req IDs |
| `assignee` | string | No | Owner |
| `sprint` | string | No | Target version |
| `started_date` | date | No | YYYY-MM-DD |
| `completed_date` | date | No | YYYY-MM-DD |
| `requirement_file` | string | No | Path to spec file |

### Phoenix Extension

Additional columns for validation taxonomy:

| Column | Type | Description |
|--------|------|-------------|
| `scope_unit` | bool | Single component test |
| `scope_integration` | bool | Multi-component test |
| `scope_system` | bool | End-to-end test |
| `technique_nominal` | bool | Happy path |
| `technique_parametric` | bool | Parameter sweep |
| `technique_monte_carlo` | bool | Random scenarios |
| `technique_stress` | bool | Edge cases |
| `env_simulation` | bool | Software-only |
| `env_hil` | bool | Hardware-in-loop |
| `env_anechoic` | bool | RF chamber |
| `env_static_field` | bool | Outdoor stationary |
| `env_dynamic_field` | bool | Outdoor moving |
| `baseline_metric` | float | Starting measurement |
| `current_metric` | float | Current measurement |
| `target_metric` | float | Goal measurement |
| `metric_unit` | string | Unit of measurement |

---

## Error Handling

### Error Categories

| Category | Code Range | Description |
|----------|------------|-------------|
| Configuration | 10-19 | Config file errors |
| Database | 20-29 | CSV/data errors |
| Validation | 30-39 | Data integrity errors |
| Integration | 40-49 | External service errors |
| Filesystem | 50-59 | File/permission errors |

### Common Errors

| Code | Error | Resolution |
|------|-------|------------|
| 10 | Config not found | Run `rtmx init` |
| 11 | Invalid config | Check rtmx.yaml syntax |
| 20 | Database not found | Check database path in config |
| 21 | Invalid CSV format | Validate CSV structure |
| 30 | Duplicate req_id | Remove duplicates |
| 31 | Invalid status | Use COMPLETE/PARTIAL/MISSING |
| 40 | GitHub auth failed | Check GITHUB_TOKEN |
| 50 | Permission denied | Check file permissions |

---

## Versioning Strategy

### What to Version Control

**Always commit:**
- `rtmx.yaml`
- `docs/rtm_database.csv`
- `docs/requirements/**/*.md`
- `.rtmx/templates/`

**Never commit (.gitignore):**
- `.rtmx/cache/`
- `.rtmx/backups/` (optional - can commit for audit trail)

### Recommended Git Workflow

```bash
# After RTM changes
git add docs/rtm_database.csv docs/requirements/
git commit -m "rtm: Update requirements [REQ-XX-NNN]"

# After config changes
git add rtmx.yaml
git commit -m "rtm: Update configuration"

# After bulk operations
git add -A
git commit -m "rtm: Bootstrap from tests"
```

### Backup Strategy

RTMX creates automatic backups before destructive operations:

```
.rtmx/backups/
├── 2024-01-15T10-30-00_remove_REQ-SW-001/
│   ├── rtm_database.csv
│   └── requirements/
├── 2024-01-16T14-00-00_reconcile/
│   └── rtm_database.csv
└── 2024-01-17T09-00-00_uninstall/
    ├── rtmx.yaml
    ├── rtm_database.csv
    └── requirements/
```

---

## Command Reference

### Core Commands

```
rtmx setup [--dry-run] [--minimal]       Complete RTMX setup (recommended)
rtmx init [--force]                      Minimal RTM initialization
rtmx status [-v|-vv|-vvv] [--json]       Show completion status
rtmx backlog [--phase N] [--view MODE]   Show incomplete requirements
rtmx health [--format FMT] [--strict]    Run integration health checks
rtmx config [--validate] [--format FMT]  Show or validate configuration
```

### Analysis Commands

```
rtmx deps [--req ID] [--category CAT]    Show dependency graph
rtmx cycles                              Detect circular dependencies
rtmx reconcile [--execute]               Fix dependency reciprocity
rtmx analyze [--format FMT] [--deep]     Discover requirements
rtmx diff BASELINE [CURRENT]             Compare RTM versions (for PRs)
```

### Integration Commands

```
rtmx from-tests [--update]               Sync from pytest markers
rtmx bootstrap [--from-tests] [--merge]  Generate from artifacts
rtmx install [--all] [--dry-run]         Install agent prompts
rtmx sync {github|jira} [--import]       Sync with external services
rtmx makefile [-o FILE]                  Generate Makefile targets
rtmx mcp-server [--port N] [--daemon]    Start MCP server
```

---

## Appendix: State Transitions

### Requirement Status Flow

```
NOT_STARTED → MISSING → PARTIAL → COMPLETE
     │           │         │
     └───────────┴─────────┴──→ (any state via manual override)
```

### Sync State Machine

```
LOCAL_ONLY ←──────────────────────→ REMOTE_ONLY
     │                                    │
     ↓                                    ↓
  SYNCED ←─── (no changes) ───────→ SYNCED
     │                                    │
     ↓                                    ↓
MODIFIED_LOCAL ←── (conflict) ──→ MODIFIED_REMOTE
     │                                    │
     └────────→ CONFLICT ←────────────────┘
```

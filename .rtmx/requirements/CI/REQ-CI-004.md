# REQ-CI-004: Screenshot Generation

## Metadata
- **Category**: CI
- **Subcategory**: Docs
- **Priority**: MEDIUM
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-010, REQ-GO-011, REQ-GO-012

## Requirement

CI pipeline shall generate terminal screenshots of key CLI commands and create PRs to update documentation assets.

## Rationale

The Python CI (ci.yml job `update-screenshots`) creates terminal screenshots of `rtmx status`, `rtmx backlog`, and `rtmx health` commands as PNG files. These are used in README and documentation. The Go CLI should generate equivalent screenshots to demonstrate its output.

## Design

### Workflow Addition (ci.yml)

New job `update-screenshots` that:
1. Depends on `test` and `lint` passing
2. Runs only on `push` to `main`
3. Builds the Go CLI
4. Creates a demo project with sample data
5. Runs commands and captures terminal output
6. Converts ANSI output to PNG using a tool (e.g., `ansi2image`, `termshot`, `vhs`)
7. Creates PR with updated screenshots

### Screenshots to Generate

| Command | Output File |
|---------|-------------|
| `rtmx status` | `docs/assets/rtmx-status.png` |
| `rtmx backlog` | `docs/assets/rtmx-backlog.png` |
| `rtmx health` | `docs/assets/rtmx-health.png` |

### Screenshot Tool Options

1. **VHS** (charmbracelet/vhs) - Go-native, declarative tape files
2. **ansi2image** (Python) - Same as rtmx uses
3. **termshot** - Simple terminal screenshot tool

Recommendation: VHS is Go-native and produces high-quality GIFs/PNGs.

## Acceptance Criteria

1. Screenshots generated from actual CLI output (not mockups)
2. PR created automatically with updated PNGs
3. Screenshots use consistent terminal dimensions and theme
4. Branch naming: `auto/update-screenshots-YYYYMMDD-HHMMSS`

## Files to Create/Modify

- `.github/workflows/ci.yml` - Add `update-screenshots` job
- `docs/assets/` - Screenshot output directory

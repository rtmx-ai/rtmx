# REQ-TUI-001: Interactive Terminal UI Framework with Bubble Tea

## Metadata
- **Category**: TUI
- **Subcategory**: Framework
- **Priority**: P0
- **Phase**: 26
- **Status**: MISSING
- **Dependencies**: (none)
- **Blocks**: REQ-TUI-002, REQ-TUI-003, REQ-TUI-004, REQ-TUI-005, REQ-TUI-006, REQ-TUI-007

## Requirement

The `rtmx tui` command shall launch an interactive terminal application
built on the Bubble Tea framework, replacing the current static one-shot
print with a full-screen application featuring keyboard navigation, multiple
views, and a persistent status bar.

## Rationale

The current `rtmx tui` command prints status counts once and exits. An
interactive TUI is the primary interface for CLI-first users who spend
their day in the terminal. Project management requires browsing, filtering,
and drilling into requirements -- operations that demand an interactive
application, not a static dump. Bubble Tea is the standard Go TUI framework
(used by gh, glow, charm) and provides the Model-View-Update architecture
needed for responsive terminal applications.

## Design

### Architecture

```
rtmx tui
  |
  +-- App Model (top-level)
  |     +-- StatusBar (bottom: mode, filter, help hints)
  |     +-- TabBar (top: Status | Backlog | Graph | Agents)
  |     +-- ActiveView (one of the tab views)
  |
  +-- Views
        +-- StatusView (REQ-TUI-002)
        +-- DetailView (REQ-TUI-003)
        +-- GraphView (REQ-TUI-004)
        +-- BacklogView (REQ-TUI-005)
        +-- AgentView (REQ-TUI-007)
```

### Key Bindings

| Key | Action |
|-----|--------|
| Tab / Shift+Tab | Switch between views |
| j/k or Up/Down | Navigate list items |
| Enter | Open detail view for selected requirement |
| / | Enter search/filter mode |
| q | Quit |
| ? | Toggle help overlay |
| r | Refresh data from database |
| 1-5 | Jump to view by number |

### Dependencies

```
github.com/charmbracelet/bubbletea  -- TUI framework
github.com/charmbracelet/bubbles    -- Standard components (list, table, textinput)
github.com/charmbracelet/lipgloss   -- Terminal styling
```

### Backward Compatibility

The `--once` flag retains the current static print behavior for scripting
and CI. The interactive mode is the new default when stdout is a terminal.
When stdout is not a terminal (piped), `--once` behavior is automatic.

### Data Loading

The TUI reads directly from the in-memory database loaded at startup. The
`r` key reloads the database from disk. No HTTP server is needed -- the TUI
is a standalone application that uses the same database and graph packages.

## Acceptance Criteria

1. `rtmx tui` launches a full-screen interactive terminal application.
2. Tab bar shows all available views and highlights the active view.
3. Status bar shows current mode, active filters, and key hints.
4. All key bindings are functional and documented in help overlay.
5. `--once` flag preserves existing static print behavior.
6. Non-TTY detection automatically falls back to `--once` mode.
7. Application starts in < 200ms for a 500-requirement database.
8. Terminal resize is handled gracefully (responsive layout).
9. No Bubble Tea dependency leaks into non-TUI packages.

## Files to Create/Modify

- `internal/tui/app.go` -- Top-level Bubble Tea model
- `internal/tui/statusbar.go` -- Status bar component
- `internal/tui/tabbar.go` -- Tab bar component
- `internal/tui/keys.go` -- Key binding definitions
- `internal/tui/styles.go` -- Lipgloss style definitions
- `internal/cmd/tui.go` -- Wire Bubble Tea app, preserve --once fallback
- `go.mod` -- Add charmbracelet dependencies

## Effort Estimate

2 weeks

## Test Strategy

- Unit test: model initialization, key handling, view switching
- Terminal detection: verify --once fallback on non-TTY
- Startup time benchmark: < 200ms with 500 requirements
- Resize handler: verify no panic on terminal resize events
- Integration test: launch TUI in PTY, send keystrokes, verify output

"""RTMX TUI - Interactive Terminal User Interface.

Provides a split-pane dashboard for browsing requirements.
Requires the textual library: pip install rtmx[tui]
"""

from __future__ import annotations

import sys
from pathlib import Path

# Detect if textual is available
_TEXTUAL_AVAILABLE = False
try:
    from textual.app import App, ComposeResult
    from textual.binding import Binding
    from textual.containers import Container, Horizontal
    from textual.widgets import DataTable, Footer, Header, Static

    _TEXTUAL_AVAILABLE = True
except ImportError:
    pass


def is_textual_available() -> bool:
    """Check if textual library is available.

    Returns:
        True if textual is installed and importable
    """
    return _TEXTUAL_AVAILABLE


# Only define TUI classes when textual is available
if _TEXTUAL_AVAILABLE:
    from rtmx.models import Requirement, RTMDatabase, Status

    class RequirementDetail(Static):
        """Widget to display requirement details."""

        def __init__(self, *args, **kwargs) -> None:
            super().__init__(*args, **kwargs)
            self._requirement = None

        def set_requirement(self, req) -> None:
            """Set the requirement to display."""
            self._requirement = req
            self._update_display()

        def _update_display(self) -> None:
            """Update the display with requirement details."""
            if self._requirement is None:
                self.update("Select a requirement to view details")
                return

            req = self._requirement
            lines = [
                f"[bold]{req.req_id}[/bold]",
                "",
                f"[dim]Status:[/dim]     {self._format_status(req.status)}",
                f"[dim]Priority:[/dim]   {req.priority.value}",
                f"[dim]Phase:[/dim]      {req.phase or 'N/A'}",
                f"[dim]Category:[/dim]   {req.category}",
                "",
                "[dim]Description:[/dim]",
                f"  {req.requirement_text}",
                "",
            ]

            if req.dependencies:
                deps = ", ".join(sorted(req.dependencies))
                lines.append(f"[dim]Dependencies:[/dim] {deps}")

            if req.blocks:
                blocks = ", ".join(sorted(req.blocks))
                lines.append(f"[dim]Blocks:[/dim] {blocks}")

            if req.effort_weeks:
                lines.append(f"[dim]Effort:[/dim] {req.effort_weeks:.1f} weeks")

            if req.notes:
                lines.append("")
                lines.append("[dim]Notes:[/dim]")
                lines.append(f"  {req.notes}")

            self.update("\n".join(lines))

        def _format_status(self, status) -> str:
            """Format status with color."""
            from rtmx.models import Status

            status_colors = {
                Status.COMPLETE: "[green]✓ COMPLETE[/green]",
                Status.PARTIAL: "[yellow]⚠ PARTIAL[/yellow]",
                Status.MISSING: "[red]✗ MISSING[/red]",
                Status.NOT_STARTED: "[dim]○ NOT STARTED[/dim]",
            }
            return status_colors.get(status, str(status.value))

    class RTMXApp(App):
        """RTMX Interactive Dashboard."""

        CSS = """
        #main-container {
            layout: horizontal;
        }

        #requirements-list {
            width: 1fr;
            border: solid green;
        }

        #detail-panel {
            width: 1fr;
            border: solid blue;
            padding: 1;
        }

        DataTable {
            height: 100%;
        }

        Footer {
            background: $surface;
        }
        """

        BINDINGS = [
            Binding("q", "quit", "Quit"),
            Binding("j", "cursor_down", "Down", show=False),
            Binding("k", "cursor_up", "Up", show=False),
            Binding("g", "cursor_top", "Top", show=False),
            Binding("G", "cursor_bottom", "Bottom", show=False),
            Binding("enter", "select_row", "Select", show=False),
            Binding("r", "refresh", "Refresh"),
        ]

        def __init__(self, rtm_csv: Path | None = None) -> None:
            super().__init__()
            self.rtm_csv = rtm_csv
            self.db: RTMDatabase | None = None
            self.requirements: list[Requirement] = []
            self._selected_index = 0

        def compose(self) -> ComposeResult:
            """Compose the application layout."""
            yield Header(show_clock=True)
            with Horizontal(id="main-container"):
                with Container(id="requirements-list"):
                    yield DataTable(id="req-table")
                with Container(id="detail-panel"):
                    yield RequirementDetail(id="detail")
            yield Footer()

        def on_mount(self) -> None:
            """Initialize the application."""
            self._load_data()
            self._populate_table()

        def _load_data(self) -> None:
            """Load requirements from database."""
            try:
                self.db = RTMDatabase.load(self.rtm_csv)
                self.requirements = list(self.db)
            except Exception as e:
                self.notify(f"Error loading database: {e}", severity="error")
                self.requirements = []

        def _populate_table(self) -> None:
            """Populate the requirements table."""
            table = self.query_one("#req-table", DataTable)
            table.clear(columns=True)

            table.add_column("Status", key="status", width=3)
            table.add_column("ID", key="id", width=15)
            table.add_column("Description", key="desc")
            table.add_column("Phase", key="phase", width=6)

            for req in self.requirements:
                status_icon = {
                    Status.COMPLETE: "✓",
                    Status.PARTIAL: "⚠",
                    Status.MISSING: "✗",
                    Status.NOT_STARTED: "○",
                }.get(req.status, "?")

                table.add_row(
                    status_icon,
                    req.req_id,
                    req.requirement_text[:50] + "..."
                    if len(req.requirement_text) > 50
                    else req.requirement_text,
                    str(req.phase) if req.phase else "-",
                    key=req.req_id,
                )

            # Update title with stats
            if self.db:
                complete = sum(1 for r in self.requirements if r.status == Status.COMPLETE)
                total = len(self.requirements)
                pct = (complete / total * 100) if total > 0 else 0
                self.title = f"RTMX Dashboard - {complete}/{total} complete ({pct:.1f}%)"
            else:
                self.title = "RTMX Dashboard"

            # Select first row if available
            if self.requirements:
                table.cursor_type = "row"
                self._update_detail(0)

        def on_data_table_row_selected(self, event: DataTable.RowSelected) -> None:
            """Handle row selection."""
            if event.row_key:
                # Find requirement by ID
                req_id = str(event.row_key.value)
                for i, req in enumerate(self.requirements):
                    if req.req_id == req_id:
                        self._update_detail(i)
                        break

        def on_data_table_row_highlighted(self, event: DataTable.RowHighlighted) -> None:
            """Handle row highlight (cursor movement)."""
            if event.row_key:
                req_id = str(event.row_key.value)
                for i, req in enumerate(self.requirements):
                    if req.req_id == req_id:
                        self._update_detail(i)
                        break

        def _update_detail(self, index: int) -> None:
            """Update the detail panel with the selected requirement."""
            if 0 <= index < len(self.requirements):
                self._selected_index = index
                detail = self.query_one("#detail", RequirementDetail)
                detail.set_requirement(self.requirements[index])

        def action_cursor_down(self) -> None:
            """Move cursor down (vim j)."""
            table = self.query_one("#req-table", DataTable)
            table.action_cursor_down()

        def action_cursor_up(self) -> None:
            """Move cursor up (vim k)."""
            table = self.query_one("#req-table", DataTable)
            table.action_cursor_up()

        def action_cursor_top(self) -> None:
            """Move cursor to top (vim g)."""
            table = self.query_one("#req-table", DataTable)
            table.move_cursor(row=0)

        def action_cursor_bottom(self) -> None:
            """Move cursor to bottom (vim G)."""
            table = self.query_one("#req-table", DataTable)
            if self.requirements:
                table.move_cursor(row=len(self.requirements) - 1)

        def action_refresh(self) -> None:
            """Refresh data from file."""
            self._load_data()
            self._populate_table()
            self.notify("Refreshed", timeout=1)

        def action_select_row(self) -> None:
            """Handle enter key on selected row."""
            # Already handled by row_highlighted, but could add modal details here
            pass


def run_tui(rtm_csv: Path | None = None) -> None:
    """Run the TUI application.

    Args:
        rtm_csv: Path to RTM CSV or None to auto-discover
    """
    from rtmx.formatting import Colors

    if not _TEXTUAL_AVAILABLE:
        print(
            f"{Colors.RED}Error: TUI requires the 'textual' library.{Colors.RESET}",
            file=sys.stderr,
        )
        print(
            f"{Colors.DIM}Install with: pip install rtmx[tui]{Colors.RESET}",
            file=sys.stderr,
        )
        sys.exit(1)

    app = RTMXApp(rtm_csv=rtm_csv)
    app.run()

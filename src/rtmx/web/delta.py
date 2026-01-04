"""Delta tracking for RTM database changes.

Provides efficient change detection to send only modified requirements
over WebSocket instead of full database refreshes.
"""

from __future__ import annotations

from pathlib import Path
from typing import Any

from rtmx.models import RTMDatabase


class StateTracker:
    """Track RTM database state for delta computation.

    Maintains the previous state of the database to compute
    differences when the file changes.
    """

    def __init__(self, rtm_csv: Path) -> None:
        """Initialize the state tracker.

        Args:
            rtm_csv: Path to the RTM database CSV file
        """
        self.rtm_csv = rtm_csv
        self.previous_state: dict[str, dict[str, Any]] | None = None

    def update(self) -> None:
        """Update the tracked state from the current database file."""
        db = RTMDatabase.load(self.rtm_csv)
        self.previous_state = {}

        for req in db:
            self.previous_state[req.req_id] = {
                "req_id": req.req_id,
                "status": req.status.value,
                "category": req.category,
                "subcategory": req.subcategory,
                "priority": req.priority.value,
                "phase": req.phase,
                "requirement_text": req.requirement_text,
                "notes": req.notes,
                "assignee": req.assignee,
            }

    def compute_delta(self) -> dict[str, list[dict[str, Any]]]:
        """Compute the delta between previous and current state.

        Returns:
            Dictionary with 'changed', 'added', 'removed' lists
        """
        if self.previous_state is None:
            return {"changed": [], "added": [], "removed": []}

        # Load current state
        db = RTMDatabase.load(self.rtm_csv)
        current_state: dict[str, dict[str, Any]] = {}

        for req in db:
            current_state[req.req_id] = {
                "req_id": req.req_id,
                "status": req.status.value,
                "category": req.category,
                "subcategory": req.subcategory,
                "priority": req.priority.value,
                "phase": req.phase,
                "requirement_text": req.requirement_text,
                "notes": req.notes,
                "assignee": req.assignee,
            }

        changed: list[dict[str, Any]] = []
        added: list[dict[str, Any]] = []
        removed: list[dict[str, Any]] = []

        # Find changed and removed requirements
        for req_id, prev_data in self.previous_state.items():
            if req_id not in current_state:
                removed.append(prev_data)
            elif current_state[req_id] != prev_data:
                # Something changed
                change_info = {
                    "req_id": req_id,
                    "old_status": prev_data["status"],
                    "new_status": current_state[req_id]["status"],
                    "changes": {},
                }
                # Detect specific field changes
                for field in prev_data:
                    if prev_data[field] != current_state[req_id].get(field):
                        change_info["changes"][field] = {
                            "old": prev_data[field],
                            "new": current_state[req_id].get(field),
                        }
                changed.append(change_info)

        # Find added requirements
        for req_id, curr_data in current_state.items():
            if req_id not in self.previous_state:
                added.append(curr_data)

        return {"changed": changed, "added": added, "removed": removed}

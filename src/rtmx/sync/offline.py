"""Offline persistence for RTMX CRDT operations.

This module provides local-first persistence for CRDT documents,
enabling offline editing with automatic sync when connectivity returns.

Features:
- Save/load CRDT document state to/from local files
- Queue pending updates when sync server unavailable
- Apply queued updates when connectivity restored
- Atomic writes to prevent corruption

Example:
    from rtmx.sync.offline import OfflineStore
    from rtmx.sync.crdt import RTMDocument

    # Initialize store
    store = OfflineStore(state_dir=Path(".rtmx/sync"))

    # Save document state
    doc = RTMDocument.from_database(db)
    store.save_state(doc)

    # Load document state
    doc = store.load_state()

    # Queue update for later sync
    store.queue_update(update_bytes)

    # Apply queued updates
    for update in store.get_pending_updates():
        doc.apply_update(update)
    store.clear_pending_updates()
"""

from __future__ import annotations

import os
import tempfile
import time
from dataclasses import dataclass, field
from pathlib import Path
from typing import TYPE_CHECKING

from rtmx.sync import require_sync

if TYPE_CHECKING:
    from rtmx.sync.crdt import RTMDocument

# Default state directory relative to project root
DEFAULT_STATE_DIR = Path(".rtmx/sync")

# File names
STATE_FILE = "state.crdt"
PENDING_DIR = "pending"


@dataclass
class OfflineStore:
    """Local persistence store for CRDT documents.

    Manages saving and loading CRDT state, plus queuing updates
    for sync when the server is unavailable.

    Attributes:
        state_dir: Directory for storing state files
        state_file: Path to the main state file
        pending_dir: Directory for pending updates
    """

    state_dir: Path = field(default_factory=lambda: DEFAULT_STATE_DIR)

    def __post_init__(self) -> None:
        """Initialize store directories."""
        require_sync()
        self.state_dir = Path(self.state_dir)
        self.state_file = self.state_dir / STATE_FILE
        self.pending_dir = self.state_dir / PENDING_DIR

    def ensure_dirs(self) -> None:
        """Create state directories if they don't exist."""
        self.state_dir.mkdir(parents=True, exist_ok=True)
        self.pending_dir.mkdir(parents=True, exist_ok=True)

    # -------------------------------------------------------------------------
    # State Persistence
    # -------------------------------------------------------------------------

    def save_state(self, doc: RTMDocument) -> Path:
        """Save document state to local file.

        Uses atomic write to prevent corruption - writes to temp file
        then renames to final location.

        Args:
            doc: RTMDocument to save

        Returns:
            Path to saved state file
        """
        self.ensure_dirs()

        # Encode full document state
        state_bytes = doc.encode_state()

        # Atomic write: write to temp, then rename
        fd, temp_path = tempfile.mkstemp(
            dir=self.state_dir,
            prefix=".state_",
            suffix=".tmp",
        )
        try:
            os.write(fd, state_bytes)
            os.fsync(fd)
        finally:
            os.close(fd)

        # Atomic rename
        os.replace(temp_path, self.state_file)

        return self.state_file

    def load_state(self) -> RTMDocument | None:
        """Load document state from local file.

        Returns:
            RTMDocument if state file exists, None otherwise
        """
        if not self.state_file.exists():
            return None

        from rtmx.sync.crdt import RTMDocument

        # Read state bytes
        state_bytes = self.state_file.read_bytes()

        # Create new document and apply state
        doc = RTMDocument()
        doc.apply_update(state_bytes)

        return doc

    def has_state(self) -> bool:
        """Check if local state file exists.

        Returns:
            True if state file exists
        """
        return self.state_file.exists()

    def delete_state(self) -> bool:
        """Delete local state file.

        Returns:
            True if deleted, False if didn't exist
        """
        if self.state_file.exists():
            self.state_file.unlink()
            return True
        return False

    # -------------------------------------------------------------------------
    # Pending Updates Queue
    # -------------------------------------------------------------------------

    def queue_update(self, update: bytes) -> Path:
        """Queue an update for later sync.

        Updates are stored as timestamped files in the pending directory.
        They are applied in order when sync resumes.

        Args:
            update: Binary CRDT update to queue

        Returns:
            Path to queued update file
        """
        self.ensure_dirs()

        # Generate unique filename with timestamp
        timestamp = int(time.time() * 1_000_000)  # Microseconds
        filename = f"{timestamp}.update"
        update_path = self.pending_dir / filename

        # Atomic write
        fd, temp_path = tempfile.mkstemp(
            dir=self.pending_dir,
            prefix=".update_",
            suffix=".tmp",
        )
        try:
            os.write(fd, update)
            os.fsync(fd)
        finally:
            os.close(fd)

        os.replace(temp_path, update_path)

        return update_path

    def get_pending_updates(self) -> list[bytes]:
        """Get all pending updates in order.

        Returns:
            List of binary CRDT updates, oldest first
        """
        if not self.pending_dir.exists():
            return []

        updates = []
        # Sort by filename (timestamp) to maintain order
        for update_file in sorted(self.pending_dir.glob("*.update")):
            updates.append(update_file.read_bytes())

        return updates

    def pending_update_count(self) -> int:
        """Get count of pending updates.

        Returns:
            Number of queued updates
        """
        if not self.pending_dir.exists():
            return 0
        return len(list(self.pending_dir.glob("*.update")))

    def clear_pending_updates(self) -> int:
        """Clear all pending updates after successful sync.

        Returns:
            Number of updates cleared
        """
        if not self.pending_dir.exists():
            return 0

        count = 0
        for update_file in self.pending_dir.glob("*.update"):
            update_file.unlink()
            count += 1

        return count

    def apply_pending_to_document(self, doc: RTMDocument) -> int:
        """Apply all pending updates to a document.

        Args:
            doc: Document to apply updates to

        Returns:
            Number of updates applied
        """
        updates = self.get_pending_updates()
        for update in updates:
            doc.apply_update(update)
        return len(updates)

    # -------------------------------------------------------------------------
    # Combined Operations
    # -------------------------------------------------------------------------

    def sync_from_csv(self, csv_path: Path) -> RTMDocument:
        """Load or create document from CSV, applying any pending updates.

        This is the main entry point for offline-first workflow:
        1. If local state exists, load it
        2. Otherwise, create from CSV
        3. Apply any pending updates
        4. Save updated state

        Args:
            csv_path: Path to CSV database

        Returns:
            RTMDocument with all local changes applied
        """
        from rtmx.sync.crdt import csv_to_crdt

        # Load existing state or create from CSV
        doc = self.load_state()
        if doc is None:
            doc = csv_to_crdt(csv_path)

        # Apply any pending updates
        self.apply_pending_to_document(doc)

        # Save updated state
        self.save_state(doc)

        return doc

    def save_and_queue(self, doc: RTMDocument, update: bytes | None = None) -> None:
        """Save document state and optionally queue an update.

        Args:
            doc: Document to save
            update: Optional update to queue for sync
        """
        self.save_state(doc)
        if update is not None:
            self.queue_update(update)


@dataclass
class SyncState:
    """Track sync status and connectivity.

    Attributes:
        is_online: Whether sync server is reachable
        last_sync: Timestamp of last successful sync
        pending_count: Number of queued updates
    """

    is_online: bool = False
    last_sync: float | None = None
    pending_count: int = 0

    def mark_synced(self) -> None:
        """Mark successful sync."""
        self.last_sync = time.time()
        self.pending_count = 0

    def mark_offline(self, pending: int = 0) -> None:
        """Mark offline with pending updates."""
        self.is_online = False
        self.pending_count = pending


__all__ = [
    "OfflineStore",
    "SyncState",
    "DEFAULT_STATE_DIR",
]

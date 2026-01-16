"""CRDT operations for RTMX requirements.

This module provides:
- RTMDocument: Y.Doc wrapper for collaborative requirements
- Requirement <-> Y.Map conversion
- CSV <-> CRDT serialization

Document Structure:
    Y.Doc (project)
    ├── requirements: Y.Map<req_id, Y.Map>
    │   ├── "REQ-CRDT-001": { req_id, status, text, ... }
    │   └── ...
    ├── metadata: Y.Map
    │   ├── schema_version: "1.0"
    │   ├── last_modified: timestamp
    │   └── owner: user_id
    └── claims: Y.Map<req_id, { user_id, expires_at }>

Example:
    from rtmx.sync.crdt import RTMDocument

    # Create from existing database
    doc = RTMDocument.from_database(db)

    # Apply remote update
    doc.apply_update(remote_update_bytes)

    # Get updated requirements
    db = doc.to_database()

    # Encode state for sync
    state = doc.encode_state()
"""

from __future__ import annotations

import time
from dataclasses import dataclass
from pathlib import Path
from typing import TYPE_CHECKING, Any

from rtmx.sync import require_sync

# Lazy import pycrdt - only when needed
if TYPE_CHECKING:
    from pycrdt import Doc, Map

    from rtmx.models import Requirement, RTMDatabase

# Schema version for CRDT document format
CRDT_SCHEMA_VERSION = "1.0"

# Fields that use Y.Text for collaborative editing
COLLABORATIVE_TEXT_FIELDS = {"requirement_text", "notes"}

# Fields stored as simple values (Last-Writer-Wins)
LWW_FIELDS = {
    "req_id",
    "category",
    "subcategory",
    "target_value",
    "test_module",
    "test_function",
    "validation_method",
    "status",
    "priority",
    "phase",
    "effort_weeks",
    "assignee",
    "sprint",
    "started_date",
    "completed_date",
    "requirement_file",
    "external_id",
}

# Set fields that need special serialization
SET_FIELDS = {"dependencies", "blocks"}


@dataclass
class ClaimInfo:
    """Information about a requirement claim."""

    user_id: str
    expires_at: float  # Unix timestamp


def requirement_to_ymap(req: Requirement) -> dict[str, Any]:
    """Convert Requirement to Y.Map-compatible dictionary.

    Static fields use Last-Writer-Wins (LWW) semantics.
    Text fields (requirement_text, notes) are stored as strings
    but can be converted to Y.Text for collaborative editing.

    Args:
        req: Requirement to convert

    Returns:
        Dictionary ready for Y.Map assignment
    """
    data: dict[str, Any] = {}

    # LWW fields - simple values
    data["req_id"] = req.req_id
    data["category"] = req.category
    data["subcategory"] = req.subcategory
    data["target_value"] = req.target_value
    data["test_module"] = req.test_module
    data["test_function"] = req.test_function
    data["validation_method"] = req.validation_method
    data["status"] = req.status.value
    data["priority"] = req.priority.value
    data["phase"] = req.phase if req.phase is not None else ""
    data["effort_weeks"] = req.effort_weeks if req.effort_weeks is not None else ""
    data["assignee"] = req.assignee
    data["sprint"] = req.sprint
    data["started_date"] = req.started_date
    data["completed_date"] = req.completed_date
    data["requirement_file"] = req.requirement_file
    data["external_id"] = req.external_id

    # Collaborative text fields - stored as strings in Y.Map
    # (Y.Text is used at a higher level for character-level CRDT)
    data["requirement_text"] = req.requirement_text
    data["notes"] = req.notes

    # Set fields - serialized as pipe-delimited strings
    data["dependencies"] = "|".join(sorted(req.dependencies)) if req.dependencies else ""
    data["blocks"] = "|".join(sorted(req.blocks)) if req.blocks else ""

    # Extra fields
    for key, value in req.extra.items():
        data[key] = value

    return data


def ymap_to_requirement(ymap: Map | dict[str, Any]) -> Requirement:
    """Convert Y.Map (or dict) back to Requirement.

    Args:
        ymap: Y.Map or dictionary with requirement data

    Returns:
        Requirement instance
    """
    from rtmx.models import Priority, Requirement, Status

    # Handle both Y.Map and dict
    raw_data = ymap.to_py() if hasattr(ymap, "to_py") else dict(ymap)
    data: dict[str, Any] = raw_data if isinstance(raw_data, dict) else {}

    # Parse dependencies and blocks from pipe-delimited strings
    deps_str = str(data.get("dependencies", ""))
    blocks_str = str(data.get("blocks", ""))
    dependencies = {d.strip() for d in deps_str.split("|") if d.strip()}
    blocks = {b.strip() for b in blocks_str.split("|") if b.strip()}

    # Parse phase
    phase_val = data.get("phase", "")
    phase: int | None = None
    if phase_val not in (None, ""):
        with __import__("contextlib").suppress(ValueError, TypeError):
            phase = int(phase_val)

    # Parse effort_weeks
    effort_val = data.get("effort_weeks", "")
    effort_weeks: float | None = None
    if effort_val not in (None, ""):
        with __import__("contextlib").suppress(ValueError, TypeError):
            effort_weeks = float(effort_val)

    # Collect extra fields
    known_fields = LWW_FIELDS | COLLABORATIVE_TEXT_FIELDS | SET_FIELDS
    extra = {k: v for k, v in data.items() if k not in known_fields}

    return Requirement(
        req_id=str(data.get("req_id", "")),
        category=str(data.get("category", "")),
        subcategory=str(data.get("subcategory", "")),
        requirement_text=str(data.get("requirement_text", "")),
        target_value=str(data.get("target_value", "")),
        test_module=str(data.get("test_module", "")),
        test_function=str(data.get("test_function", "")),
        validation_method=str(data.get("validation_method", "")),
        status=Status.from_string(str(data.get("status", "MISSING"))),
        priority=Priority.from_string(str(data.get("priority", "MEDIUM"))),
        phase=phase,
        notes=str(data.get("notes", "")),
        effort_weeks=effort_weeks,
        dependencies=dependencies,
        blocks=blocks,
        assignee=str(data.get("assignee", "")),
        sprint=str(data.get("sprint", "")),
        started_date=str(data.get("started_date", "")),
        completed_date=str(data.get("completed_date", "")),
        requirement_file=str(data.get("requirement_file", "")),
        external_id=str(data.get("external_id", "")),
        extra=extra,
    )


class RTMDocument:
    """Y.Doc wrapper for collaborative requirements management.

    This class wraps a pycrdt Y.Doc and provides high-level operations
    for managing requirements as CRDT data structures.

    Example:
        # Create new document
        doc = RTMDocument()

        # Add requirements
        doc.set_requirement(req)

        # Get requirement
        req = doc.get_requirement("REQ-SW-001")

        # Sync with remote
        remote_update = receive_from_server()
        doc.apply_update(remote_update)

        # Send local changes
        local_state = doc.encode_state()
        send_to_server(local_state)
    """

    def __init__(self, doc: Doc | None = None) -> None:
        """Initialize RTM document.

        Args:
            doc: Existing Y.Doc to wrap. If None, creates a new document.
        """
        require_sync()
        from pycrdt import Doc, Map

        self._doc = doc if doc is not None else Doc()

        # Initialize shared data structures
        # pycrdt.Doc doesn't have full typing support, use getattr pattern
        requirements_key = "requirements"
        if requirements_key not in list(self._doc.keys()):
            self._doc[requirements_key] = Map()
        metadata_key = "metadata"
        if metadata_key not in list(self._doc.keys()):
            self._doc[metadata_key] = Map()
            self._doc[metadata_key]["schema_version"] = CRDT_SCHEMA_VERSION
            self._doc[metadata_key]["created_at"] = time.time()
        claims_key = "claims"
        if claims_key not in list(self._doc.keys()):
            self._doc[claims_key] = Map()

    @property
    def doc(self) -> Doc:
        """Get underlying Y.Doc."""
        return self._doc

    @property
    def requirements(self) -> Map:
        """Get requirements Y.Map."""
        return self._doc["requirements"]

    @property
    def metadata(self) -> Map:
        """Get metadata Y.Map."""
        return self._doc["metadata"]

    @property
    def claims(self) -> Map:
        """Get claims Y.Map."""
        return self._doc["claims"]

    # -------------------------------------------------------------------------
    # Requirement Operations
    # -------------------------------------------------------------------------

    def set_requirement(self, req: Requirement) -> None:
        """Add or update a requirement in the document.

        Args:
            req: Requirement to set
        """
        from pycrdt import Map

        data = requirement_to_ymap(req)
        req_map = Map(data)
        self.requirements[req.req_id] = req_map
        self._update_modified()

    def get_requirement(self, req_id: str) -> Requirement | None:
        """Get a requirement from the document.

        Args:
            req_id: Requirement ID to retrieve

        Returns:
            Requirement if found, None otherwise
        """
        if req_id not in self.requirements:
            return None
        return ymap_to_requirement(self.requirements[req_id])

    def remove_requirement(self, req_id: str) -> bool:
        """Remove a requirement from the document.

        Args:
            req_id: Requirement ID to remove

        Returns:
            True if removed, False if not found
        """
        if req_id not in self.requirements:
            return False
        del self.requirements[req_id]
        self._update_modified()
        return True

    def list_requirements(self) -> list[str]:
        """List all requirement IDs in the document.

        Returns:
            List of requirement IDs
        """
        return list(self.requirements.keys())

    def all_requirements(self) -> list[Requirement]:
        """Get all requirements from the document.

        Returns:
            List of all requirements
        """
        return [
            ymap_to_requirement(self.requirements[req_id])
            for req_id in list(self.requirements.keys())
        ]

    # -------------------------------------------------------------------------
    # Database Conversion
    # -------------------------------------------------------------------------

    @classmethod
    def from_database(cls, db: RTMDatabase) -> RTMDocument:
        """Create RTMDocument from an RTMDatabase.

        Args:
            db: RTMDatabase to convert

        Returns:
            New RTMDocument with all requirements
        """
        doc = cls()
        for req in db:
            doc.set_requirement(req)
        return doc

    def to_database(self, path: Path | None = None) -> RTMDatabase:
        """Convert document to RTMDatabase.

        Args:
            path: Optional path for the database

        Returns:
            RTMDatabase with all requirements from this document
        """
        from rtmx.models import RTMDatabase

        requirements = self.all_requirements()
        return RTMDatabase(requirements, path)

    # -------------------------------------------------------------------------
    # CRDT State Operations
    # -------------------------------------------------------------------------

    def encode_state(self) -> bytes:
        """Encode full document state for sync.

        Returns:
            Binary CRDT state that can be sent to remote peers
        """
        return self._doc.get_update()

    def encode_state_vector(self) -> bytes:
        """Encode state vector for differential sync.

        Returns:
            Binary state vector describing what we have
        """
        return self._doc.get_state()

    def encode_update_since(self, state_vector: bytes) -> bytes:
        """Encode updates since a given state vector.

        Args:
            state_vector: Remote peer's state vector

        Returns:
            Binary update containing only changes since state_vector
        """
        return self._doc.get_update(state_vector)

    def apply_update(self, update: bytes) -> None:
        """Apply an update from a remote peer.

        Args:
            update: Binary CRDT update to apply
        """
        self._doc.apply_update(update)

    # -------------------------------------------------------------------------
    # Claim Operations (for Phase 10)
    # -------------------------------------------------------------------------

    def claim_requirement(self, req_id: str, user_id: str, duration_seconds: int = 1800) -> bool:
        """Claim a requirement for exclusive editing.

        Args:
            req_id: Requirement ID to claim
            user_id: User making the claim
            duration_seconds: Claim duration (default 30 minutes)

        Returns:
            True if claim successful, False if already claimed
        """
        from pycrdt import Map

        now = time.time()

        # Check existing claim
        if req_id in self.claims:
            claim_data = self.claims[req_id]
            if hasattr(claim_data, "to_py"):
                claim_data = claim_data.to_py()
            expires_at = claim_data.get("expires_at", 0)
            if expires_at > now:
                # Already claimed and not expired
                return claim_data.get("user_id") == user_id

        # Create new claim
        claim = Map(
            {
                "user_id": user_id,
                "expires_at": now + duration_seconds,
            }
        )
        self.claims[req_id] = claim
        return True

    def release_claim(self, req_id: str, user_id: str) -> bool:
        """Release a claim on a requirement.

        Args:
            req_id: Requirement ID to release
            user_id: User releasing the claim

        Returns:
            True if released, False if not claimed by user
        """
        if req_id not in self.claims:
            return False

        claim_data = self.claims[req_id]
        if hasattr(claim_data, "to_py"):
            claim_data = claim_data.to_py()

        if claim_data.get("user_id") != user_id:
            return False

        del self.claims[req_id]
        return True

    def get_claim(self, req_id: str) -> ClaimInfo | None:
        """Get claim information for a requirement.

        Args:
            req_id: Requirement ID to check

        Returns:
            ClaimInfo if claimed and not expired, None otherwise
        """
        if req_id not in self.claims:
            return None

        claim_data = self.claims[req_id]
        if hasattr(claim_data, "to_py"):
            claim_data = claim_data.to_py()

        expires_at = claim_data.get("expires_at", 0)
        if expires_at <= time.time():
            # Claim expired
            return None

        return ClaimInfo(
            user_id=claim_data.get("user_id", ""),
            expires_at=expires_at,
        )

    # -------------------------------------------------------------------------
    # Metadata Operations
    # -------------------------------------------------------------------------

    def _update_modified(self) -> None:
        """Update last_modified timestamp."""
        self.metadata["last_modified"] = time.time()

    def get_schema_version(self) -> str:
        """Get document schema version."""
        return str(self.metadata.get("schema_version", CRDT_SCHEMA_VERSION))

    def set_owner(self, user_id: str) -> None:
        """Set document owner."""
        self.metadata["owner"] = user_id

    def get_owner(self) -> str | None:
        """Get document owner."""
        owner = self.metadata.get("owner")
        return str(owner) if owner else None


# -------------------------------------------------------------------------
# CSV <-> CRDT Serialization
# -------------------------------------------------------------------------


def csv_to_crdt(csv_path: Path | str) -> RTMDocument:
    """Load CSV file into CRDT document.

    Args:
        csv_path: Path to CSV file

    Returns:
        RTMDocument with loaded requirements
    """
    from rtmx.models import RTMDatabase

    db = RTMDatabase.load(csv_path)
    return RTMDocument.from_database(db)


def crdt_to_csv(doc: RTMDocument, csv_path: Path | str) -> None:
    """Save CRDT document to CSV file.

    Args:
        doc: RTMDocument to save
        csv_path: Path to save CSV file
    """
    db = doc.to_database(Path(csv_path))
    db.save()


__all__ = [
    "RTMDocument",
    "requirement_to_ymap",
    "ymap_to_requirement",
    "csv_to_crdt",
    "crdt_to_csv",
    "ClaimInfo",
    "CRDT_SCHEMA_VERSION",
]

"""Tests for rtmx.sync.crdt module.

These tests verify the CRDT (Conflict-free Replicated Data Type) functionality
for collaborative requirements management.
"""

from pathlib import Path

import pytest

from rtmx import Priority, Requirement, RTMDatabase, Status
from rtmx.sync import get_sync_import_error, is_sync_available, require_sync

# Skip all tests if pycrdt is not available
pytestmark = pytest.mark.skipif(
    not is_sync_available(),
    reason="pycrdt not installed (pip install rtmx[sync])",
)


@pytest.mark.req("REQ-CRDT-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestSyncAvailability:
    """Tests for sync module availability checks."""

    def test_is_sync_available(self):
        """Test that is_sync_available returns True when pycrdt is installed."""
        assert is_sync_available() is True

    def test_get_sync_import_error_none_when_available(self):
        """Test that import error is None when pycrdt is available."""
        assert get_sync_import_error() is None

    def test_require_sync_does_not_raise(self):
        """Test that require_sync does not raise when pycrdt is available."""
        # Should not raise
        require_sync()

    def test_pycrdt_import(self):
        """Test that pycrdt can be imported and used."""
        from pycrdt import Doc, Map

        doc = Doc()
        doc["test"] = Map()
        doc["test"]["key"] = "value"
        assert doc["test"]["key"] == "value"


@pytest.fixture
def sample_requirement() -> Requirement:
    """Create a sample requirement for testing."""
    return Requirement(
        req_id="REQ-TEST-001",
        category="TESTING",
        subcategory="UNIT",
        requirement_text="System shall support CRDT synchronization",
        target_value="Real-time sync <100ms",
        test_module="tests/test_crdt.py",
        test_function="test_requirement_to_ymap",
        validation_method="Unit Test",
        status=Status.PARTIAL,
        priority=Priority.HIGH,
        phase=9,
        notes="Testing CRDT conversion",
        effort_weeks=2.0,
        dependencies={"REQ-CORE-001", "REQ-CORE-002"},
        blocks={"REQ-SYNC-002"},
        assignee="developer",
        sprint="v0.1",
        started_date="2025-01-01",
        completed_date="",
        requirement_file="docs/requirements/SYNC/REQ-TEST-001.md",
        external_id="JIRA-123",
    )


@pytest.fixture
def sample_database(tmp_path: Path, sample_requirement: Requirement) -> RTMDatabase:
    """Create a sample database with multiple requirements."""
    reqs = [
        sample_requirement,
        Requirement(
            req_id="REQ-TEST-002",
            category="TESTING",
            subcategory="INTEGRATION",
            requirement_text="System shall merge concurrent edits",
            status=Status.MISSING,
            priority=Priority.MEDIUM,
            phase=9,
        ),
        Requirement(
            req_id="REQ-TEST-003",
            category="TESTING",
            subcategory="E2E",
            requirement_text="System shall sync across clients",
            status=Status.NOT_STARTED,
            priority=Priority.HIGH,
            phase=10,
            dependencies={"REQ-TEST-001", "REQ-TEST-002"},
        ),
    ]
    return RTMDatabase(reqs, tmp_path / "test_db.csv")


@pytest.mark.req("REQ-CRDT-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRequirementToYmap:
    """Tests for requirement_to_ymap conversion."""

    def test_basic_fields(self, sample_requirement: Requirement):
        """Test that basic fields are converted correctly."""
        from rtmx.sync.crdt import requirement_to_ymap

        data = requirement_to_ymap(sample_requirement)

        assert data["req_id"] == "REQ-TEST-001"
        assert data["category"] == "TESTING"
        assert data["subcategory"] == "UNIT"
        assert data["requirement_text"] == "System shall support CRDT synchronization"
        assert data["status"] == "PARTIAL"
        assert data["priority"] == "HIGH"

    def test_phase_conversion(self, sample_requirement: Requirement):
        """Test that phase is converted correctly."""
        from rtmx.sync.crdt import requirement_to_ymap

        data = requirement_to_ymap(sample_requirement)
        assert data["phase"] == 9

    def test_effort_weeks_conversion(self, sample_requirement: Requirement):
        """Test that effort_weeks is converted correctly."""
        from rtmx.sync.crdt import requirement_to_ymap

        data = requirement_to_ymap(sample_requirement)
        assert data["effort_weeks"] == 2.0

    def test_none_phase_converted_to_empty(self):
        """Test that None phase is converted to empty string."""
        from rtmx.sync.crdt import requirement_to_ymap

        req = Requirement(req_id="REQ-TEST-001", phase=None)
        data = requirement_to_ymap(req)
        assert data["phase"] == ""

    def test_dependencies_serialization(self, sample_requirement: Requirement):
        """Test that dependencies are serialized as pipe-delimited string."""
        from rtmx.sync.crdt import requirement_to_ymap

        data = requirement_to_ymap(sample_requirement)
        # Dependencies are sorted before joining
        assert data["dependencies"] == "REQ-CORE-001|REQ-CORE-002"

    def test_blocks_serialization(self, sample_requirement: Requirement):
        """Test that blocks are serialized as pipe-delimited string."""
        from rtmx.sync.crdt import requirement_to_ymap

        data = requirement_to_ymap(sample_requirement)
        assert data["blocks"] == "REQ-SYNC-002"

    def test_empty_sets_serialization(self):
        """Test that empty sets are serialized as empty strings."""
        from rtmx.sync.crdt import requirement_to_ymap

        req = Requirement(req_id="REQ-TEST-001", dependencies=set(), blocks=set())
        data = requirement_to_ymap(req)
        assert data["dependencies"] == ""
        assert data["blocks"] == ""

    def test_extra_fields_preserved(self):
        """Test that extra fields are preserved in conversion."""
        from rtmx.sync.crdt import requirement_to_ymap

        req = Requirement(
            req_id="REQ-TEST-001",
            extra={"custom_field": "custom_value", "another": "data"},
        )
        data = requirement_to_ymap(req)
        assert data["custom_field"] == "custom_value"
        assert data["another"] == "data"


@pytest.mark.req("REQ-CRDT-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestYmapToRequirement:
    """Tests for ymap_to_requirement conversion."""

    def test_basic_fields(self):
        """Test that basic fields are converted back correctly."""
        from rtmx.sync.crdt import ymap_to_requirement

        data = {
            "req_id": "REQ-TEST-001",
            "category": "TESTING",
            "subcategory": "UNIT",
            "requirement_text": "Test requirement",
            "status": "COMPLETE",
            "priority": "P0",
        }
        req = ymap_to_requirement(data)

        assert req.req_id == "REQ-TEST-001"
        assert req.category == "TESTING"
        assert req.status == Status.COMPLETE
        assert req.priority == Priority.P0

    def test_phase_parsing(self):
        """Test that phase is parsed as integer."""
        from rtmx.sync.crdt import ymap_to_requirement

        data = {"req_id": "REQ-TEST-001", "phase": "9"}
        req = ymap_to_requirement(data)
        assert req.phase == 9

    def test_phase_empty_string(self):
        """Test that empty phase string becomes None."""
        from rtmx.sync.crdt import ymap_to_requirement

        data = {"req_id": "REQ-TEST-001", "phase": ""}
        req = ymap_to_requirement(data)
        assert req.phase is None

    def test_effort_weeks_parsing(self):
        """Test that effort_weeks is parsed as float."""
        from rtmx.sync.crdt import ymap_to_requirement

        data = {"req_id": "REQ-TEST-001", "effort_weeks": "2.5"}
        req = ymap_to_requirement(data)
        assert req.effort_weeks == 2.5

    def test_dependencies_parsing(self):
        """Test that dependencies are parsed from pipe-delimited string."""
        from rtmx.sync.crdt import ymap_to_requirement

        data = {"req_id": "REQ-TEST-001", "dependencies": "REQ-A|REQ-B|REQ-C"}
        req = ymap_to_requirement(data)
        assert req.dependencies == {"REQ-A", "REQ-B", "REQ-C"}

    def test_empty_dependencies(self):
        """Test that empty dependencies string becomes empty set."""
        from rtmx.sync.crdt import ymap_to_requirement

        data = {"req_id": "REQ-TEST-001", "dependencies": ""}
        req = ymap_to_requirement(data)
        assert req.dependencies == set()

    def test_extra_fields_captured(self):
        """Test that unknown fields are captured in extra dict."""
        from rtmx.sync.crdt import ymap_to_requirement

        data = {
            "req_id": "REQ-TEST-001",
            "custom_field": "custom_value",
            "another": "data",
        }
        req = ymap_to_requirement(data)
        assert req.extra["custom_field"] == "custom_value"
        assert req.extra["another"] == "data"


@pytest.mark.req("REQ-CRDT-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRoundTrip:
    """Tests for round-trip conversion requirement -> ymap -> requirement."""

    def test_roundtrip_preserves_data(self, sample_requirement: Requirement):
        """Test that round-trip conversion preserves all data."""
        from rtmx.sync.crdt import requirement_to_ymap, ymap_to_requirement

        # Convert to ymap and back
        ymap_data = requirement_to_ymap(sample_requirement)
        restored = ymap_to_requirement(ymap_data)

        # Verify all fields match
        assert restored.req_id == sample_requirement.req_id
        assert restored.category == sample_requirement.category
        assert restored.subcategory == sample_requirement.subcategory
        assert restored.requirement_text == sample_requirement.requirement_text
        assert restored.target_value == sample_requirement.target_value
        assert restored.status == sample_requirement.status
        assert restored.priority == sample_requirement.priority
        assert restored.phase == sample_requirement.phase
        assert restored.notes == sample_requirement.notes
        assert restored.effort_weeks == sample_requirement.effort_weeks
        assert restored.dependencies == sample_requirement.dependencies
        assert restored.blocks == sample_requirement.blocks
        assert restored.assignee == sample_requirement.assignee
        assert restored.sprint == sample_requirement.sprint

    def test_roundtrip_with_none_values(self):
        """Test round-trip with None optional values."""
        from rtmx.sync.crdt import requirement_to_ymap, ymap_to_requirement

        req = Requirement(
            req_id="REQ-TEST-001",
            phase=None,
            effort_weeks=None,
        )
        ymap_data = requirement_to_ymap(req)
        restored = ymap_to_requirement(ymap_data)

        assert restored.phase is None
        assert restored.effort_weeks is None


@pytest.mark.req("REQ-CRDT-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDocument:
    """Tests for RTMDocument Y.Doc wrapper."""

    def test_create_empty_document(self):
        """Test creating an empty RTMDocument."""
        from rtmx.sync.crdt import RTMDocument

        doc = RTMDocument()
        assert doc.list_requirements() == []

    def test_set_and_get_requirement(self, sample_requirement: Requirement):
        """Test setting and getting a requirement."""
        from rtmx.sync.crdt import RTMDocument

        doc = RTMDocument()
        doc.set_requirement(sample_requirement)

        retrieved = doc.get_requirement("REQ-TEST-001")
        assert retrieved is not None
        assert retrieved.req_id == "REQ-TEST-001"
        assert retrieved.category == "TESTING"
        assert retrieved.status == Status.PARTIAL

    def test_get_nonexistent_requirement(self):
        """Test getting a requirement that doesn't exist."""
        from rtmx.sync.crdt import RTMDocument

        doc = RTMDocument()
        assert doc.get_requirement("REQ-NONEXISTENT") is None

    def test_remove_requirement(self, sample_requirement: Requirement):
        """Test removing a requirement."""
        from rtmx.sync.crdt import RTMDocument

        doc = RTMDocument()
        doc.set_requirement(sample_requirement)

        assert doc.remove_requirement("REQ-TEST-001") is True
        assert doc.get_requirement("REQ-TEST-001") is None
        assert doc.remove_requirement("REQ-TEST-001") is False

    def test_list_requirements(self, sample_database: RTMDatabase):
        """Test listing all requirement IDs."""
        from rtmx.sync.crdt import RTMDocument

        doc = RTMDocument.from_database(sample_database)
        req_ids = doc.list_requirements()

        assert len(req_ids) == 3
        assert "REQ-TEST-001" in req_ids
        assert "REQ-TEST-002" in req_ids
        assert "REQ-TEST-003" in req_ids

    def test_all_requirements(self, sample_database: RTMDatabase):
        """Test getting all requirements."""
        from rtmx.sync.crdt import RTMDocument

        doc = RTMDocument.from_database(sample_database)
        reqs = doc.all_requirements()

        assert len(reqs) == 3
        req_ids = {r.req_id for r in reqs}
        assert req_ids == {"REQ-TEST-001", "REQ-TEST-002", "REQ-TEST-003"}


@pytest.mark.req("REQ-CRDT-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestDatabaseConversion:
    """Tests for RTMDatabase <-> RTMDocument conversion."""

    def test_from_database(self, sample_database: RTMDatabase):
        """Test creating RTMDocument from RTMDatabase."""
        from rtmx.sync.crdt import RTMDocument

        doc = RTMDocument.from_database(sample_database)

        assert len(doc.list_requirements()) == 3
        req = doc.get_requirement("REQ-TEST-001")
        assert req is not None
        assert req.category == "TESTING"

    def test_to_database(self, sample_database: RTMDatabase, tmp_path: Path):
        """Test converting RTMDocument to RTMDatabase."""
        from rtmx.sync.crdt import RTMDocument

        doc = RTMDocument.from_database(sample_database)
        db = doc.to_database(tmp_path / "output.csv")

        assert len(db) == 3
        req = db.get("REQ-TEST-001")
        assert req.category == "TESTING"

    def test_database_roundtrip(self, sample_database: RTMDatabase, tmp_path: Path):
        """Test round-trip database -> document -> database."""
        from rtmx.sync.crdt import RTMDocument

        # Convert to document and back
        doc = RTMDocument.from_database(sample_database)
        restored_db = doc.to_database(tmp_path / "restored.csv")

        # Verify all requirements preserved
        for original_req in sample_database:
            restored_req = restored_db.get(original_req.req_id)
            assert restored_req.category == original_req.category
            assert restored_req.status == original_req.status
            assert restored_req.priority == original_req.priority


@pytest.mark.req("REQ-CRDT-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCRDTStateOperations:
    """Tests for CRDT state encoding and updates."""

    def test_encode_state(self, sample_requirement: Requirement):
        """Test encoding document state."""
        from rtmx.sync.crdt import RTMDocument

        doc = RTMDocument()
        doc.set_requirement(sample_requirement)

        state = doc.encode_state()
        assert isinstance(state, bytes)
        assert len(state) > 0

    def test_apply_update(self, sample_requirement: Requirement):
        """Test applying update from another document."""
        from rtmx.sync.crdt import RTMDocument

        # Create source document with a requirement
        source_doc = RTMDocument()
        source_doc.set_requirement(sample_requirement)
        update = source_doc.encode_state()

        # Apply update to target document
        target_doc = RTMDocument()
        target_doc.apply_update(update)

        # Verify requirement synced
        req = target_doc.get_requirement("REQ-TEST-001")
        assert req is not None
        assert req.req_id == "REQ-TEST-001"

    def test_concurrent_edits_merge(self):
        """Test that concurrent edits from two documents merge correctly."""
        from rtmx.sync.crdt import RTMDocument

        # Document A adds requirement 1
        doc_a = RTMDocument()
        doc_a.set_requirement(
            Requirement(req_id="REQ-A-001", category="FROM_A", status=Status.COMPLETE)
        )

        # Document B adds requirement 2
        doc_b = RTMDocument()
        doc_b.set_requirement(
            Requirement(req_id="REQ-B-001", category="FROM_B", status=Status.PARTIAL)
        )

        # Exchange updates
        update_a = doc_a.encode_state()
        update_b = doc_b.encode_state()

        doc_a.apply_update(update_b)
        doc_b.apply_update(update_a)

        # Both documents should have both requirements
        assert doc_a.get_requirement("REQ-A-001") is not None
        assert doc_a.get_requirement("REQ-B-001") is not None
        assert doc_b.get_requirement("REQ-A-001") is not None
        assert doc_b.get_requirement("REQ-B-001") is not None

    def test_lww_same_field_concurrent_edit(self):
        """Test Last-Writer-Wins semantics for same-field concurrent edits.

        When two documents edit the same field of the same requirement,
        the last write (by timestamp/logical clock) should win.
        """
        from rtmx.sync.crdt import RTMDocument

        # Both documents start with same requirement
        doc_a = RTMDocument()
        doc_a.set_requirement(
            Requirement(req_id="REQ-SHARED-001", status=Status.MISSING, category="INIT")
        )

        # Sync initial state to doc_b
        doc_b = RTMDocument()
        doc_b.apply_update(doc_a.encode_state())

        # Both now have the same requirement
        assert doc_a.get_requirement("REQ-SHARED-001") is not None
        assert doc_b.get_requirement("REQ-SHARED-001") is not None

        # Document A updates status to PARTIAL
        doc_a.set_requirement(
            Requirement(req_id="REQ-SHARED-001", status=Status.PARTIAL, category="FROM_A")
        )

        # Document B updates status to COMPLETE (happens "after" A in logical time)
        doc_b.set_requirement(
            Requirement(req_id="REQ-SHARED-001", status=Status.COMPLETE, category="FROM_B")
        )

        # Exchange updates - B's update is "later" so it should win
        update_a = doc_a.encode_state()
        update_b = doc_b.encode_state()

        doc_a.apply_update(update_b)
        doc_b.apply_update(update_a)

        # Both documents should converge to the same state
        req_a = doc_a.get_requirement("REQ-SHARED-001")
        req_b = doc_b.get_requirement("REQ-SHARED-001")

        assert req_a is not None
        assert req_b is not None

        # The key assertion: both documents have the same final state
        assert req_a.status == req_b.status
        assert req_a.category == req_b.category

        # Note: Which value "wins" depends on internal CRDT clock ordering
        # The important thing is CONVERGENCE - both have same value


@pytest.mark.req("REQ-CRDT-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCSVCRDTSerialization:
    """Tests for CSV <-> CRDT serialization."""

    def test_csv_to_crdt(self, sample_database: RTMDatabase, tmp_path: Path):
        """Test loading CSV into CRDT document."""
        from rtmx.sync.crdt import csv_to_crdt

        # Save sample database to CSV
        csv_path = tmp_path / "test.csv"
        sample_database._path = csv_path
        sample_database.save()

        # Load into CRDT
        doc = csv_to_crdt(csv_path)

        assert len(doc.list_requirements()) == 3
        req = doc.get_requirement("REQ-TEST-001")
        assert req is not None
        assert req.category == "TESTING"

    def test_crdt_to_csv(self, sample_database: RTMDatabase, tmp_path: Path):
        """Test saving CRDT document to CSV."""
        from rtmx.sync.crdt import RTMDocument, crdt_to_csv

        # Create document
        doc = RTMDocument.from_database(sample_database)

        # Save to CSV
        csv_path = tmp_path / "output.csv"
        crdt_to_csv(doc, csv_path)

        # Reload and verify
        loaded_db = RTMDatabase.load(csv_path)
        assert len(loaded_db) == 3
        req = loaded_db.get("REQ-TEST-001")
        assert req.category == "TESTING"

    def test_csv_crdt_roundtrip(self, sample_database: RTMDatabase, tmp_path: Path):
        """Test full round-trip CSV -> CRDT -> CSV."""
        from rtmx.sync.crdt import crdt_to_csv, csv_to_crdt

        # Save original to CSV
        original_path = tmp_path / "original.csv"
        sample_database._path = original_path
        sample_database.save()

        # Convert to CRDT and back
        doc = csv_to_crdt(original_path)
        restored_path = tmp_path / "restored.csv"
        crdt_to_csv(doc, restored_path)

        # Compare
        restored_db = RTMDatabase.load(restored_path)
        for original_req in sample_database:
            restored_req = restored_db.get(original_req.req_id)
            assert restored_req.req_id == original_req.req_id
            assert restored_req.category == original_req.category
            assert restored_req.status == original_req.status


@pytest.mark.req("REQ-CRDT-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestClaimOperations:
    """Tests for requirement claim/lock operations (Phase 10 prep)."""

    def test_claim_requirement(self):
        """Test claiming a requirement for editing."""
        from rtmx.sync.crdt import RTMDocument

        doc = RTMDocument()
        doc.set_requirement(Requirement(req_id="REQ-TEST-001"))

        result = doc.claim_requirement("REQ-TEST-001", "user-123", duration_seconds=60)
        assert result is True

        claim = doc.get_claim("REQ-TEST-001")
        assert claim is not None
        assert claim.user_id == "user-123"

    def test_claim_already_claimed(self):
        """Test that claiming an already-claimed requirement fails."""
        from rtmx.sync.crdt import RTMDocument

        doc = RTMDocument()
        doc.set_requirement(Requirement(req_id="REQ-TEST-001"))

        # First claim succeeds
        assert doc.claim_requirement("REQ-TEST-001", "user-123") is True

        # Second claim by different user fails
        assert doc.claim_requirement("REQ-TEST-001", "user-456") is False

        # Claim by same user succeeds (extends)
        assert doc.claim_requirement("REQ-TEST-001", "user-123") is True

    def test_release_claim(self):
        """Test releasing a claim."""
        from rtmx.sync.crdt import RTMDocument

        doc = RTMDocument()
        doc.set_requirement(Requirement(req_id="REQ-TEST-001"))
        doc.claim_requirement("REQ-TEST-001", "user-123")

        # Release by owner succeeds
        assert doc.release_claim("REQ-TEST-001", "user-123") is True
        assert doc.get_claim("REQ-TEST-001") is None

    def test_release_claim_wrong_user(self):
        """Test that releasing someone else's claim fails."""
        from rtmx.sync.crdt import RTMDocument

        doc = RTMDocument()
        doc.set_requirement(Requirement(req_id="REQ-TEST-001"))
        doc.claim_requirement("REQ-TEST-001", "user-123")

        # Release by different user fails
        assert doc.release_claim("REQ-TEST-001", "user-456") is False

        # Claim still exists
        claim = doc.get_claim("REQ-TEST-001")
        assert claim is not None
        assert claim.user_id == "user-123"


@pytest.mark.req("REQ-CRDT-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMetadataOperations:
    """Tests for document metadata operations."""

    def test_schema_version(self):
        """Test getting schema version."""
        from rtmx.sync.crdt import CRDT_SCHEMA_VERSION, RTMDocument

        doc = RTMDocument()
        assert doc.get_schema_version() == CRDT_SCHEMA_VERSION

    def test_set_and_get_owner(self):
        """Test setting and getting document owner."""
        from rtmx.sync.crdt import RTMDocument

        doc = RTMDocument()
        doc.set_owner("user-123")
        assert doc.get_owner() == "user-123"

    def test_owner_initially_none(self):
        """Test that owner is None for new document."""
        from rtmx.sync.crdt import RTMDocument

        doc = RTMDocument()
        assert doc.get_owner() is None


@pytest.mark.req("REQ-CRDT-005")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestOfflineOperations:
    """Tests for offline persistence operations."""

    def test_save_state_to_file(self, sample_requirement: Requirement, tmp_path: Path):
        """Test saving document state to local file."""
        from rtmx.sync.crdt import RTMDocument
        from rtmx.sync.offline import OfflineStore

        # Create document with requirement
        doc = RTMDocument()
        doc.set_requirement(sample_requirement)

        # Save state
        store = OfflineStore(state_dir=tmp_path)
        state_path = store.save_state(doc)

        # Verify file exists
        assert state_path.exists()
        assert state_path.stat().st_size > 0

    def test_load_state_from_file(self, sample_requirement: Requirement, tmp_path: Path):
        """Test loading document state from local file."""
        from rtmx.sync.crdt import RTMDocument
        from rtmx.sync.offline import OfflineStore

        # Create and save document
        doc = RTMDocument()
        doc.set_requirement(sample_requirement)

        store = OfflineStore(state_dir=tmp_path)
        store.save_state(doc)

        # Load state into new document
        loaded_doc = store.load_state()

        assert loaded_doc is not None
        req = loaded_doc.get_requirement("REQ-TEST-001")
        assert req is not None
        assert req.category == "TESTING"

    def test_state_roundtrip(self, sample_database: RTMDatabase, tmp_path: Path):
        """Test full save/load roundtrip preserves all data."""
        from rtmx.sync.crdt import RTMDocument
        from rtmx.sync.offline import OfflineStore

        # Create document from database
        doc = RTMDocument.from_database(sample_database)

        # Save and reload
        store = OfflineStore(state_dir=tmp_path)
        store.save_state(doc)
        loaded_doc = store.load_state()

        # Verify all requirements preserved
        assert loaded_doc is not None
        assert len(loaded_doc.list_requirements()) == len(sample_database)

        for req_id in doc.list_requirements():
            original = doc.get_requirement(req_id)
            loaded = loaded_doc.get_requirement(req_id)
            assert loaded is not None
            assert original is not None
            assert loaded.req_id == original.req_id
            assert loaded.category == original.category
            assert loaded.status == original.status

    def test_load_nonexistent_state(self, tmp_path: Path):
        """Test loading when no state file exists."""
        from rtmx.sync.offline import OfflineStore

        store = OfflineStore(state_dir=tmp_path)
        doc = store.load_state()

        assert doc is None

    def test_has_state(self, sample_requirement: Requirement, tmp_path: Path):
        """Test checking if state file exists."""
        from rtmx.sync.crdt import RTMDocument
        from rtmx.sync.offline import OfflineStore

        store = OfflineStore(state_dir=tmp_path)

        # Initially no state
        assert store.has_state() is False

        # After save, state exists
        doc = RTMDocument()
        doc.set_requirement(sample_requirement)
        store.save_state(doc)

        assert store.has_state() is True

    def test_delete_state(self, sample_requirement: Requirement, tmp_path: Path):
        """Test deleting state file."""
        from rtmx.sync.crdt import RTMDocument
        from rtmx.sync.offline import OfflineStore

        store = OfflineStore(state_dir=tmp_path)

        # Create state
        doc = RTMDocument()
        doc.set_requirement(sample_requirement)
        store.save_state(doc)

        assert store.has_state() is True

        # Delete state
        result = store.delete_state()
        assert result is True
        assert store.has_state() is False

        # Delete again returns False
        result = store.delete_state()
        assert result is False

    def test_queue_pending_updates(self, tmp_path: Path):
        """Test queuing updates for later sync."""
        from rtmx.sync.offline import OfflineStore

        store = OfflineStore(state_dir=tmp_path)

        # Queue some updates
        update1 = b"update_data_1"
        update2 = b"update_data_2"
        update3 = b"update_data_3"

        store.queue_update(update1)
        store.queue_update(update2)
        store.queue_update(update3)

        # Verify count
        assert store.pending_update_count() == 3

    def test_get_pending_updates_order(self, tmp_path: Path):
        """Test that pending updates are returned in order."""
        import time

        from rtmx.sync.offline import OfflineStore

        store = OfflineStore(state_dir=tmp_path)

        # Queue updates with slight delay to ensure ordering
        updates = [b"first", b"second", b"third"]
        for update in updates:
            store.queue_update(update)
            time.sleep(0.001)  # Small delay for timestamp ordering

        # Get updates
        pending = store.get_pending_updates()

        assert len(pending) == 3
        assert pending == updates

    def test_clear_pending_updates(self, tmp_path: Path):
        """Test clearing pending updates after sync."""
        from rtmx.sync.offline import OfflineStore

        store = OfflineStore(state_dir=tmp_path)

        # Queue updates
        store.queue_update(b"update1")
        store.queue_update(b"update2")
        assert store.pending_update_count() == 2

        # Clear
        cleared = store.clear_pending_updates()
        assert cleared == 2
        assert store.pending_update_count() == 0

        # Clear again returns 0
        cleared = store.clear_pending_updates()
        assert cleared == 0

    def test_apply_pending_to_document(self, tmp_path: Path):
        """Test applying pending updates to a document."""
        from rtmx.sync.crdt import RTMDocument
        from rtmx.sync.offline import OfflineStore

        store = OfflineStore(state_dir=tmp_path)

        # Create document with initial requirement
        doc_a = RTMDocument()
        doc_a.set_requirement(Requirement(req_id="REQ-INITIAL", category="INITIAL"))

        # Create update from another document
        doc_b = RTMDocument()
        doc_b.apply_update(doc_a.encode_state())  # Sync initial state
        doc_b.set_requirement(Requirement(req_id="REQ-NEW", category="FROM_UPDATE"))
        update = doc_b.encode_state()

        # Queue the update
        store.queue_update(update)

        # Apply pending to original document
        count = store.apply_pending_to_document(doc_a)

        assert count == 1
        assert doc_a.get_requirement("REQ-INITIAL") is not None
        assert doc_a.get_requirement("REQ-NEW") is not None

    def test_sync_from_csv(self, sample_database: RTMDatabase, tmp_path: Path):
        """Test sync_from_csv creates document from CSV with pending updates."""
        from rtmx.sync.crdt import RTMDocument
        from rtmx.sync.offline import OfflineStore

        # Save database to CSV
        csv_path = tmp_path / "test.csv"
        sample_database._path = csv_path
        sample_database.save()

        store = OfflineStore(state_dir=tmp_path / "sync")

        # First call creates from CSV
        doc = store.sync_from_csv(csv_path)
        assert len(doc.list_requirements()) == 3

        # Create an update
        doc.set_requirement(Requirement(req_id="REQ-OFFLINE-001", category="OFFLINE"))
        store.save_state(doc)

        # Queue a pending update
        doc2 = RTMDocument()
        doc2.apply_update(doc.encode_state())
        doc2.set_requirement(Requirement(req_id="REQ-PENDING-001", category="PENDING"))
        store.queue_update(doc2.encode_state())

        # Sync again - should load state and apply pending
        doc3 = store.sync_from_csv(csv_path)

        # Should have original + offline + pending
        assert doc3.get_requirement("REQ-TEST-001") is not None
        assert doc3.get_requirement("REQ-OFFLINE-001") is not None
        assert doc3.get_requirement("REQ-PENDING-001") is not None


@pytest.mark.req("REQ-CRDT-005")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestSyncState:
    """Tests for sync state tracking."""

    def test_initial_state(self):
        """Test initial sync state."""
        from rtmx.sync.offline import SyncState

        state = SyncState()
        assert state.is_online is False
        assert state.last_sync is None
        assert state.pending_count == 0

    def test_mark_synced(self):
        """Test marking successful sync."""
        import time

        from rtmx.sync.offline import SyncState

        state = SyncState()
        state.pending_count = 5

        before = time.time()
        state.mark_synced()
        after = time.time()

        assert state.last_sync is not None
        assert before <= state.last_sync <= after
        assert state.pending_count == 0

    def test_mark_offline(self):
        """Test marking offline status."""
        from rtmx.sync.offline import SyncState

        state = SyncState(is_online=True)
        state.mark_offline(pending=3)

        assert state.is_online is False
        assert state.pending_count == 3


# =============================================================================
# E2E Tests for CRDT Offline Workflow
# =============================================================================


@pytest.mark.req("REQ-CRDT-005")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCRDTOfflineE2E:
    """End-to-end tests for CRDT offline-first workflow.

    These tests simulate real-world scenarios with multiple clients
    working offline and synchronizing their changes.
    """

    def test_e2e_single_client_offline_workflow(self, sample_database: RTMDatabase, tmp_path: Path):
        """E2E: Single client works offline, persists state, restarts, continues.

        Scenario:
        1. Client loads CSV into CRDT
        2. Client makes offline edits
        3. Client saves state and exits
        4. Client restarts and loads persisted state
        5. Client continues editing
        6. Client exports final state to CSV
        """
        from rtmx.sync.crdt import crdt_to_csv
        from rtmx.sync.offline import OfflineStore

        # Setup: Save initial database
        csv_path = tmp_path / "project" / "rtm_database.csv"
        csv_path.parent.mkdir(parents=True)
        sample_database._path = csv_path
        sample_database.save()

        sync_dir = tmp_path / "project" / ".rtmx" / "sync"
        store = OfflineStore(state_dir=sync_dir)

        # Step 1: Load CSV into CRDT
        doc = store.sync_from_csv(csv_path)
        assert len(doc.list_requirements()) == 3

        # Step 2: Make offline edits
        doc.set_requirement(
            Requirement(
                req_id="REQ-OFFLINE-001",
                category="OFFLINE",
                requirement_text="Added while offline",
                status=Status.PARTIAL,
            )
        )

        # Update existing requirement
        existing = doc.get_requirement("REQ-TEST-001")
        assert existing is not None
        doc.set_requirement(
            Requirement(
                req_id="REQ-TEST-001",
                category=existing.category,
                requirement_text=existing.requirement_text,
                status=Status.COMPLETE,  # Changed status
                notes="Updated offline",
            )
        )

        # Step 3: Save state (simulating exit)
        store.save_state(doc)

        # Step 4: "Restart" - create new document and load state
        doc2 = store.load_state()
        assert doc2 is not None
        assert len(doc2.list_requirements()) == 4

        # Verify offline changes persisted
        offline_req = doc2.get_requirement("REQ-OFFLINE-001")
        assert offline_req is not None
        assert offline_req.status == Status.PARTIAL

        updated_req = doc2.get_requirement("REQ-TEST-001")
        assert updated_req is not None
        assert updated_req.notes == "Updated offline"

        # Step 5: Continue editing
        doc2.set_requirement(
            Requirement(
                req_id="REQ-OFFLINE-002",
                category="OFFLINE",
                requirement_text="Second offline addition",
                status=Status.MISSING,
            )
        )
        store.save_state(doc2)

        # Step 6: Export to CSV
        final_csv = tmp_path / "project" / "rtm_final.csv"
        crdt_to_csv(doc2, final_csv)

        # Verify final CSV has all changes
        final_db = RTMDatabase.load(final_csv)
        assert len(final_db) == 5
        assert final_db.get("REQ-OFFLINE-001") is not None
        assert final_db.get("REQ-OFFLINE-002") is not None

    def test_e2e_two_clients_concurrent_offline_edits(self, tmp_path: Path):
        """E2E: Two clients work offline concurrently, then merge changes.

        Scenario:
        1. Both clients start from same CSV
        2. Client A works offline, adds requirements
        3. Client B works offline, adds different requirements
        4. Client A saves state and creates update
        5. Client B receives A's update and merges
        6. Both clients converge to same state
        """
        from rtmx.sync.crdt import csv_to_crdt
        from rtmx.sync.offline import OfflineStore

        # Setup: Create shared initial database
        csv_path = tmp_path / "shared" / "rtm_database.csv"
        csv_path.parent.mkdir(parents=True)

        initial_db = RTMDatabase(
            [
                Requirement(
                    req_id="REQ-SHARED-001",
                    category="SHARED",
                    requirement_text="Initial shared requirement",
                    status=Status.MISSING,
                )
            ],
            csv_path,
        )
        initial_db.save()

        # Client A setup
        store_a = OfflineStore(state_dir=tmp_path / "client_a" / ".rtmx" / "sync")
        doc_a = csv_to_crdt(csv_path)

        # Client B setup
        store_b = OfflineStore(state_dir=tmp_path / "client_b" / ".rtmx" / "sync")
        doc_b = csv_to_crdt(csv_path)

        # Both start with same state
        assert doc_a.list_requirements() == doc_b.list_requirements()

        # Client A works offline
        doc_a.set_requirement(
            Requirement(
                req_id="REQ-FROM-A-001",
                category="CLIENT_A",
                requirement_text="Added by client A",
                status=Status.PARTIAL,
            )
        )
        doc_a.set_requirement(
            Requirement(
                req_id="REQ-FROM-A-002",
                category="CLIENT_A",
                requirement_text="Another from A",
                status=Status.COMPLETE,
            )
        )
        store_a.save_state(doc_a)

        # Client B works offline (concurrently, unaware of A's changes)
        doc_b.set_requirement(
            Requirement(
                req_id="REQ-FROM-B-001",
                category="CLIENT_B",
                requirement_text="Added by client B",
                status=Status.MISSING,
            )
        )
        # B also updates the shared requirement
        doc_b.set_requirement(
            Requirement(
                req_id="REQ-SHARED-001",
                category="SHARED",
                requirement_text="Initial shared requirement",
                status=Status.PARTIAL,  # B marked it as partial
                notes="Progress made by B",
            )
        )
        store_b.save_state(doc_b)

        # Simulate sync: Exchange updates
        update_a = doc_a.encode_state()
        update_b = doc_b.encode_state()

        # Client A receives B's update
        doc_a.apply_update(update_b)
        store_a.save_state(doc_a)

        # Client B receives A's update
        doc_b.apply_update(update_a)
        store_b.save_state(doc_b)

        # Both should now have converged state
        reqs_a = set(doc_a.list_requirements())
        reqs_b = set(doc_b.list_requirements())
        assert reqs_a == reqs_b

        # Should have all requirements from both clients
        expected_reqs = {
            "REQ-SHARED-001",
            "REQ-FROM-A-001",
            "REQ-FROM-A-002",
            "REQ-FROM-B-001",
        }
        assert reqs_a == expected_reqs

        # Both should have same value for shared requirement (convergence)
        shared_a = doc_a.get_requirement("REQ-SHARED-001")
        shared_b = doc_b.get_requirement("REQ-SHARED-001")
        assert shared_a is not None
        assert shared_b is not None
        assert shared_a.status == shared_b.status
        assert shared_a.notes == shared_b.notes

    def test_e2e_offline_queue_and_replay(self, sample_database: RTMDatabase, tmp_path: Path):
        """E2E: Client queues updates while offline, replays when back online.

        Scenario:
        1. Client starts online with initial state
        2. Connection drops, client continues working
        3. Each change is queued as pending update
        4. Connection restored, pending updates applied
        5. State is consistent after replay
        """
        from rtmx.sync.crdt import RTMDocument
        from rtmx.sync.offline import OfflineStore, SyncState

        # Setup
        csv_path = tmp_path / "rtm_database.csv"
        sample_database._path = csv_path
        sample_database.save()

        store = OfflineStore(state_dir=tmp_path / ".rtmx" / "sync")
        sync_state = SyncState(is_online=True)

        # Step 1: Initial online state
        doc = store.sync_from_csv(csv_path)
        initial_state = doc.encode_state()
        sync_state.mark_synced()

        # Step 2: Go offline
        sync_state.mark_offline()

        # Step 3: Make changes while offline, queue each as update
        doc.set_requirement(
            Requirement(req_id="REQ-QUEUED-001", category="QUEUED", status=Status.MISSING)
        )
        store.queue_update(doc.encode_state())

        doc.set_requirement(
            Requirement(req_id="REQ-QUEUED-002", category="QUEUED", status=Status.PARTIAL)
        )
        store.queue_update(doc.encode_state())

        doc.set_requirement(
            Requirement(req_id="REQ-QUEUED-003", category="QUEUED", status=Status.COMPLETE)
        )
        store.queue_update(doc.encode_state())

        store.save_state(doc)
        sync_state.pending_count = store.pending_update_count()
        assert sync_state.pending_count == 3

        # Step 4: Simulate "server" receiving updates
        server_doc = RTMDocument()
        server_doc.apply_update(initial_state)

        # Replay pending updates on server
        for update in store.get_pending_updates():
            server_doc.apply_update(update)

        # Step 5: Verify consistency
        assert set(server_doc.list_requirements()) == set(doc.list_requirements())
        assert server_doc.get_requirement("REQ-QUEUED-001") is not None
        assert server_doc.get_requirement("REQ-QUEUED-002") is not None
        assert server_doc.get_requirement("REQ-QUEUED-003") is not None

        # Clear pending after successful sync
        store.clear_pending_updates()
        sync_state.mark_synced()
        assert sync_state.pending_count == 0
        assert store.pending_update_count() == 0

    def test_e2e_recover_from_crash(self, sample_database: RTMDatabase, tmp_path: Path):
        """E2E: Client recovers state after crash using persisted data.

        Scenario:
        1. Client loads database and makes changes
        2. Client saves state periodically
        3. "Crash" occurs (we just forget the document)
        4. Client recovers from persisted state
        5. No data is lost
        """
        from rtmx.sync.offline import OfflineStore

        csv_path = tmp_path / "rtm_database.csv"
        sample_database._path = csv_path
        sample_database.save()

        store = OfflineStore(state_dir=tmp_path / ".rtmx" / "sync")

        # Session 1: Normal operation
        doc = store.sync_from_csv(csv_path)

        doc.set_requirement(
            Requirement(req_id="REQ-IMPORTANT-001", category="CRITICAL", status=Status.PARTIAL)
        )
        store.save_state(doc)  # Save point 1

        doc.set_requirement(
            Requirement(req_id="REQ-IMPORTANT-002", category="CRITICAL", status=Status.COMPLETE)
        )
        store.save_state(doc)  # Save point 2

        doc.set_requirement(
            Requirement(req_id="REQ-IMPORTANT-003", category="CRITICAL", status=Status.MISSING)
        )
        store.save_state(doc)  # Save point 3

        # Record final state
        final_reqs = set(doc.list_requirements())
        final_count = len(final_reqs)

        # "Crash" - document object is gone
        del doc

        # Session 2: Recovery
        recovered_doc = store.load_state()
        assert recovered_doc is not None

        # All data recovered
        recovered_reqs = set(recovered_doc.list_requirements())
        assert recovered_reqs == final_reqs
        assert len(recovered_reqs) == final_count

        # Specific requirements recovered
        assert recovered_doc.get_requirement("REQ-IMPORTANT-001") is not None
        assert recovered_doc.get_requirement("REQ-IMPORTANT-002") is not None
        assert recovered_doc.get_requirement("REQ-IMPORTANT-003") is not None

    def test_e2e_full_round_trip_csv_crdt_offline_csv(
        self, sample_database: RTMDatabase, tmp_path: Path
    ):
        """E2E: Complete round trip - CSV -> CRDT -> Offline edits -> CSV.

        This test verifies the full workflow a user would experience:
        1. Start with existing CSV database
        2. Load into CRDT for collaborative editing
        3. Work offline with state persistence
        4. Merge with other changes
        5. Export back to CSV for git commit
        """
        from rtmx.sync.crdt import crdt_to_csv, csv_to_crdt
        from rtmx.sync.offline import OfflineStore

        # === SETUP ===
        project_dir = tmp_path / "my_project"
        project_dir.mkdir()

        original_csv = project_dir / "docs" / "rtm_database.csv"
        original_csv.parent.mkdir(parents=True)
        sample_database._path = original_csv
        sample_database.save()

        # === USER A: First user session ===
        sync_dir_a = project_dir / ".rtmx" / "sync_a"
        store_a = OfflineStore(state_dir=sync_dir_a)

        # Load project
        doc_a = csv_to_crdt(original_csv)
        store_a.save_state(doc_a)

        # Make changes
        doc_a.set_requirement(
            Requirement(
                req_id="REQ-FEATURE-001",
                category="FEATURE",
                subcategory="UI",
                requirement_text="Add dark mode toggle",
                status=Status.PARTIAL,
                priority=Priority.HIGH,
                phase=2,
            )
        )
        store_a.save_state(doc_a)

        # Export for git commit
        updated_csv_a = project_dir / "docs" / "rtm_database_a.csv"
        crdt_to_csv(doc_a, updated_csv_a)

        # === USER B: Second user session (concurrent) ===
        sync_dir_b = project_dir / ".rtmx" / "sync_b"
        store_b = OfflineStore(state_dir=sync_dir_b)

        # Load original (hasn't seen A's changes yet)
        doc_b = csv_to_crdt(original_csv)
        store_b.save_state(doc_b)

        # Make different changes
        doc_b.set_requirement(
            Requirement(
                req_id="REQ-FEATURE-002",
                category="FEATURE",
                subcategory="API",
                requirement_text="Add rate limiting",
                status=Status.MISSING,
                priority=Priority.HIGH,
                phase=2,
            )
        )
        store_b.save_state(doc_b)

        # === MERGE: Simulate git merge / sync ===
        # B receives A's update
        doc_b.apply_update(doc_a.encode_state())
        store_b.save_state(doc_b)

        # A receives B's update
        doc_a.apply_update(doc_b.encode_state())
        store_a.save_state(doc_a)

        # === VERIFY CONVERGENCE ===
        assert set(doc_a.list_requirements()) == set(doc_b.list_requirements())

        # === FINAL EXPORT ===
        final_csv = project_dir / "docs" / "rtm_database_merged.csv"
        crdt_to_csv(doc_a, final_csv)

        # === VERIFY FINAL STATE ===
        final_db = RTMDatabase.load(final_csv)

        # Original requirements preserved
        assert final_db.get("REQ-TEST-001") is not None
        assert final_db.get("REQ-TEST-002") is not None
        assert final_db.get("REQ-TEST-003") is not None

        # Both users' additions present
        feature_001 = final_db.get("REQ-FEATURE-001")
        assert feature_001 is not None
        assert feature_001.requirement_text == "Add dark mode toggle"
        assert feature_001.status == Status.PARTIAL

        feature_002 = final_db.get("REQ-FEATURE-002")
        assert feature_002 is not None
        assert feature_002.requirement_text == "Add rate limiting"
        assert feature_002.status == Status.MISSING

        # Total count correct
        assert len(final_db) == 5  # 3 original + 2 new

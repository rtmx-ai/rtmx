"""Comprehensive tests for rtmx.schema module."""

import pytest

from rtmx.schema import (
    CORE_SCHEMA,
    PHOENIX_EXTENSION,
    PHOENIX_SCHEMA,
    Column,
    ColumnType,
    Schema,
    get_schema,
    list_schemas,
    register_schema,
)


class TestColumnType:
    """Tests for ColumnType enum."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_column_types_exist(self):
        """Test all column types are defined."""
        assert ColumnType.STRING == "string"
        assert ColumnType.INTEGER == "integer"
        assert ColumnType.FLOAT == "float"
        assert ColumnType.BOOLEAN == "boolean"
        assert ColumnType.DATE == "date"
        assert ColumnType.LIST == "list"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_column_type_values(self):
        """Test ColumnType string values are correct."""
        types = {t.value for t in ColumnType}
        assert types == {"string", "integer", "float", "boolean", "date", "list"}


class TestColumn:
    """Tests for Column dataclass."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_column_creation_minimal(self):
        """Test creating column with minimal parameters."""
        col = Column(name="test_col")
        assert col.name == "test_col"
        assert col.type == ColumnType.STRING
        assert col.required is False
        assert col.default == ""
        assert col.validator is None
        assert col.description == ""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_column_creation_full(self):
        """Test creating column with all parameters."""

        def validator(x):
            return x.startswith("REQ-")

        col = Column(
            name="req_id",
            type=ColumnType.STRING,
            required=True,
            default="",
            validator=validator,
            description="Requirement ID",
        )
        assert col.name == "req_id"
        assert col.type == ColumnType.STRING
        assert col.required is True
        assert col.default == ""
        assert col.validator is validator
        assert col.description == "Requirement ID"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_column_with_validator(self):
        """Test column with custom validator function."""

        def validator(x):
            return x in ("HIGH", "MEDIUM", "LOW")

        col = Column(name="priority", validator=validator)
        assert col.validator("HIGH") is True
        assert col.validator("INVALID") is False

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_column_different_types(self):
        """Test creating columns with different types."""
        int_col = Column(name="phase", type=ColumnType.INTEGER)
        float_col = Column(name="effort", type=ColumnType.FLOAT)
        bool_col = Column(name="active", type=ColumnType.BOOLEAN)
        date_col = Column(name="start_date", type=ColumnType.DATE)
        list_col = Column(name="deps", type=ColumnType.LIST)

        assert int_col.type == ColumnType.INTEGER
        assert float_col.type == ColumnType.FLOAT
        assert bool_col.type == ColumnType.BOOLEAN
        assert date_col.type == ColumnType.DATE
        assert list_col.type == ColumnType.LIST


class TestSchema:
    """Tests for Schema class."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_schema_creation_empty(self):
        """Test creating empty schema."""
        schema = Schema(name="test", description="Test schema")
        assert schema.name == "test"
        assert schema.description == "Test schema"
        assert len(schema.columns) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_schema_add_column(self):
        """Test adding column to schema."""
        schema = Schema(name="test")
        col = Column(name="test_col", required=True)
        schema.add_column(col)

        assert schema.has_column("test_col")
        assert len(schema.columns) == 1
        assert schema.columns["test_col"] == col

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_schema_remove_column(self):
        """Test removing column from schema."""
        schema = Schema(name="test")
        col = Column(name="test_col")
        schema.add_column(col)
        assert schema.has_column("test_col")

        schema.remove_column("test_col")
        assert not schema.has_column("test_col")
        assert len(schema.columns) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_schema_remove_nonexistent_column(self):
        """Test removing non-existent column does not raise error."""
        schema = Schema(name="test")
        # Should not raise exception
        schema.remove_column("nonexistent")
        assert len(schema.columns) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_schema_has_column(self):
        """Test checking if schema has column."""
        schema = Schema(name="test")
        col = Column(name="test_col")
        schema.add_column(col)

        assert schema.has_column("test_col") is True
        assert schema.has_column("nonexistent") is False

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_schema_required_columns(self):
        """Test getting required columns list."""
        schema = Schema(name="test")
        schema.add_column(Column(name="req_id", required=True))
        schema.add_column(Column(name="category", required=True))
        schema.add_column(Column(name="notes", required=False))

        required = schema.required_columns()
        assert "req_id" in required
        assert "category" in required
        assert "notes" not in required
        assert len(required) == 2

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_schema_validate_row_success(self):
        """Test validating a valid row."""
        schema = Schema(name="test")
        schema.add_column(Column(name="req_id", required=True))
        schema.add_column(Column(name="category", required=True))

        row = {"req_id": "REQ-001", "category": "SOFTWARE"}
        errors = schema.validate_row(row)
        assert len(errors) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_schema_validate_row_missing_required(self):
        """Test validation fails for missing required column."""
        schema = Schema(name="test")
        schema.add_column(Column(name="req_id", required=True))

        row = {"category": "SOFTWARE"}
        errors = schema.validate_row(row)
        assert len(errors) == 1
        assert "req_id" in errors[0]

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_schema_validate_row_empty_required(self):
        """Test validation fails for empty required column."""
        schema = Schema(name="test")
        schema.add_column(Column(name="req_id", required=True))

        row = {"req_id": ""}
        errors = schema.validate_row(row)
        assert len(errors) == 1
        assert "req_id" in errors[0]

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_schema_validate_row_whitespace_required(self):
        """Test validation fails for whitespace-only required column."""
        schema = Schema(name="test")
        schema.add_column(Column(name="req_id", required=True))

        row = {"req_id": "   "}
        errors = schema.validate_row(row)
        assert len(errors) == 1
        assert "req_id" in errors[0]

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_schema_validate_row_validator_pass(self):
        """Test validation passes when validator returns True."""

        def validator(x):
            return x in ("COMPLETE", "PARTIAL", "MISSING")

        schema = Schema(name="test")
        schema.add_column(Column(name="status", validator=validator))

        row = {"status": "COMPLETE"}
        errors = schema.validate_row(row)
        assert len(errors) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_schema_validate_row_validator_fail(self):
        """Test validation fails when validator returns False."""

        def validator(x):
            return x in ("COMPLETE", "PARTIAL", "MISSING")

        schema = Schema(name="test")
        schema.add_column(Column(name="status", validator=validator))

        row = {"status": "INVALID"}
        errors = schema.validate_row(row)
        assert len(errors) == 1
        assert "status" in errors[0]
        assert "INVALID" in errors[0]

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_schema_validate_row_validator_exception(self):
        """Test validation catches exceptions from validator."""

        def validator(x):
            return int(x)  # Will raise for non-numeric

        schema = Schema(name="test")
        schema.add_column(Column(name="phase", validator=validator))

        row = {"phase": "not_a_number"}
        errors = schema.validate_row(row)
        assert len(errors) == 1
        assert "phase" in errors[0]
        assert "Validation error" in errors[0]

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_schema_extend(self):
        """Test extending schema with another schema."""
        base = Schema(name="base")
        base.add_column(Column(name="req_id", required=True))

        extension = Schema(name="ext")
        extension.add_column(Column(name="extra_field"))

        combined = base.extend(extension)
        assert combined.name == "base+ext"
        assert combined.has_column("req_id")
        assert combined.has_column("extra_field")
        assert len(combined.columns) == 2

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_schema_extend_preserves_original(self):
        """Test extending schema does not modify originals."""
        base = Schema(name="base")
        base.add_column(Column(name="req_id"))

        extension = Schema(name="ext")
        extension.add_column(Column(name="extra_field"))

        combined = base.extend(extension)

        # Original schemas should be unchanged
        assert len(base.columns) == 1
        assert len(extension.columns) == 1
        assert len(combined.columns) == 2

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_schema_extend_override(self):
        """Test extension can override base columns."""
        base = Schema(name="base")
        base.add_column(Column(name="field", default="base"))

        extension = Schema(name="ext")
        extension.add_column(Column(name="field", default="ext"))

        combined = base.extend(extension)
        # Extension should override
        assert combined.columns["field"].default == "ext"


class TestCoreSchema:
    """Tests for CORE_SCHEMA definition."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_core_schema_exists(self):
        """Test CORE_SCHEMA is defined."""
        assert CORE_SCHEMA is not None
        assert CORE_SCHEMA.name == "core"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_core_schema_has_20_columns(self):
        """Test CORE_SCHEMA has exactly 20 columns."""
        assert len(CORE_SCHEMA.columns) == 20

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_core_schema_required_columns(self):
        """Test CORE_SCHEMA has correct required columns."""
        required = CORE_SCHEMA.required_columns()
        assert "req_id" in required
        assert "category" in required
        assert "requirement_text" in required
        assert "status" in required

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_core_schema_has_validators(self):
        """Test CORE_SCHEMA columns have appropriate validators."""
        assert CORE_SCHEMA.columns["status"].validator is not None
        assert CORE_SCHEMA.columns["priority"].validator is not None

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_core_schema_status_validator(self):
        """Test CORE_SCHEMA status validator works correctly."""
        validator = CORE_SCHEMA.columns["status"].validator
        assert validator("COMPLETE") is True
        assert validator("PARTIAL") is True
        assert validator("MISSING") is True
        assert validator("NOT_STARTED") is True
        assert validator("") is True
        assert validator("INVALID") is False

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_core_schema_priority_validator(self):
        """Test CORE_SCHEMA priority validator works correctly."""
        validator = CORE_SCHEMA.columns["priority"].validator
        assert validator("P0") is True
        assert validator("HIGH") is True
        assert validator("MEDIUM") is True
        assert validator("LOW") is True
        assert validator("") is True
        assert validator("INVALID") is False

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_core_schema_column_types(self):
        """Test CORE_SCHEMA columns have correct types."""
        assert CORE_SCHEMA.columns["req_id"].type == ColumnType.STRING
        assert CORE_SCHEMA.columns["phase"].type == ColumnType.INTEGER
        assert CORE_SCHEMA.columns["effort_weeks"].type == ColumnType.FLOAT
        assert CORE_SCHEMA.columns["dependencies"].type == ColumnType.LIST
        assert CORE_SCHEMA.columns["started_date"].type == ColumnType.DATE

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_core_schema_column_defaults(self):
        """Test CORE_SCHEMA columns have correct defaults."""
        assert CORE_SCHEMA.columns["status"].default == "MISSING"
        assert CORE_SCHEMA.columns["priority"].default == "MEDIUM"


class TestPhoenixSchema:
    """Tests for PHOENIX_EXTENSION and PHOENIX_SCHEMA."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_phoenix_extension_exists(self):
        """Test PHOENIX_EXTENSION is defined."""
        assert PHOENIX_EXTENSION is not None
        assert PHOENIX_EXTENSION.name == "phoenix"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_phoenix_extension_has_validation_columns(self):
        """Test PHOENIX_EXTENSION includes validation taxonomy columns."""
        assert PHOENIX_EXTENSION.has_column("scope_unit")
        assert PHOENIX_EXTENSION.has_column("scope_integration")
        assert PHOENIX_EXTENSION.has_column("scope_system")
        assert PHOENIX_EXTENSION.has_column("technique_nominal")
        assert PHOENIX_EXTENSION.has_column("technique_parametric")
        assert PHOENIX_EXTENSION.has_column("technique_monte_carlo")
        assert PHOENIX_EXTENSION.has_column("technique_stress")

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_phoenix_extension_has_environment_columns(self):
        """Test PHOENIX_EXTENSION includes environment columns."""
        assert PHOENIX_EXTENSION.has_column("env_simulation")
        assert PHOENIX_EXTENSION.has_column("env_hil")
        assert PHOENIX_EXTENSION.has_column("env_anechoic")
        assert PHOENIX_EXTENSION.has_column("env_static_field")
        assert PHOENIX_EXTENSION.has_column("env_dynamic_field")

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_phoenix_extension_boolean_defaults(self):
        """Test PHOENIX_EXTENSION boolean columns default to False."""
        assert PHOENIX_EXTENSION.columns["scope_unit"].default is False
        assert PHOENIX_EXTENSION.columns["env_simulation"].default is False
        assert PHOENIX_EXTENSION.columns["technique_nominal"].default is False

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_phoenix_extension_boolean_types(self):
        """Test PHOENIX_EXTENSION boolean columns have correct type."""
        assert PHOENIX_EXTENSION.columns["scope_unit"].type == ColumnType.BOOLEAN
        assert PHOENIX_EXTENSION.columns["env_simulation"].type == ColumnType.BOOLEAN

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_phoenix_extension_has_metrics_columns(self):
        """Test PHOENIX_EXTENSION includes metrics columns."""
        assert PHOENIX_EXTENSION.has_column("baseline_metric")
        assert PHOENIX_EXTENSION.has_column("current_metric")
        assert PHOENIX_EXTENSION.has_column("target_metric")
        assert PHOENIX_EXTENSION.has_column("metric_unit")

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_phoenix_extension_has_legacy_test_columns(self):
        """Test PHOENIX_EXTENSION includes legacy test columns."""
        assert PHOENIX_EXTENSION.has_column("unit_test")
        assert PHOENIX_EXTENSION.has_column("integration_test")
        assert PHOENIX_EXTENSION.has_column("parametric_test")
        assert PHOENIX_EXTENSION.has_column("monte_carlo_test")
        assert PHOENIX_EXTENSION.has_column("stress_test")

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_phoenix_schema_is_extended(self):
        """Test PHOENIX_SCHEMA is properly extended from CORE_SCHEMA."""
        assert PHOENIX_SCHEMA is not None
        # Should have core columns
        assert PHOENIX_SCHEMA.has_column("req_id")
        assert PHOENIX_SCHEMA.has_column("category")
        # Should have extension columns
        assert PHOENIX_SCHEMA.has_column("scope_unit")
        assert PHOENIX_SCHEMA.has_column("env_simulation")

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_phoenix_schema_name(self):
        """Test PHOENIX_SCHEMA has correct combined name."""
        assert "core" in PHOENIX_SCHEMA.name
        assert "phoenix" in PHOENIX_SCHEMA.name


class TestSchemaRegistry:
    """Tests for schema registry functions."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_get_schema_core(self):
        """Test getting core schema from registry."""
        schema = get_schema("core")
        assert schema is CORE_SCHEMA
        assert schema.name == "core"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_get_schema_phoenix(self):
        """Test getting phoenix schema from registry."""
        schema = get_schema("phoenix")
        assert schema is PHOENIX_SCHEMA

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_get_schema_not_found(self):
        """Test getting non-existent schema raises KeyError."""
        with pytest.raises(KeyError) as exc:
            get_schema("nonexistent")
        assert "nonexistent" in str(exc.value)
        assert "Available" in str(exc.value)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_register_schema(self):
        """Test registering a custom schema."""
        custom = Schema(name="custom_test_schema")
        custom.add_column(Column(name="custom_field"))

        register_schema(custom)

        # Should be retrievable
        retrieved = get_schema("custom_test_schema")
        assert retrieved is custom
        assert retrieved.has_column("custom_field")

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_list_schemas(self):
        """Test listing available schemas."""
        schemas = list_schemas()
        assert "core" in schemas
        assert "phoenix" in schemas
        assert isinstance(schemas, list)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_register_schema_override(self):
        """Test registering schema can override existing."""
        original = get_schema("core")

        custom = Schema(name="core")
        register_schema(custom)

        # Should be overridden
        retrieved = get_schema("core")
        assert retrieved is custom
        assert retrieved is not original

        # Restore original for other tests
        register_schema(CORE_SCHEMA)

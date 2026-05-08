package schema

import (
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestSchemaLoad(t *testing.T) {
	rtmx.Req(t, "REQ-PLUGIN-005")

	t.Run("create_schema_with_columns", func(t *testing.T) {
		s := New("test", []Column{
			{Name: "id", Type: TypeString, Required: true},
			{Name: "count", Type: TypeInt},
			{Name: "score", Type: TypeFloat},
			{Name: "active", Type: TypeBool},
			{Name: "created", Type: TypeDate},
			{Name: "status", Type: TypeEnum, EnumValues: []string{"OPEN", "CLOSED"}},
			{Name: "tags", Type: TypeSet, EnumValues: []string{"A", "B", "C"}},
		})

		if s.Name != "test" {
			t.Errorf("Name = %q, want test", s.Name)
		}
		if len(s.Columns) != 7 {
			t.Errorf("Columns = %d, want 7", len(s.Columns))
		}
		names := s.ColumnNames()
		if names[0] != "id" || names[6] != "tags" {
			t.Errorf("ColumnNames = %v", names)
		}
	})

	t.Run("has_column", func(t *testing.T) {
		s := New("test", []Column{
			{Name: "id", Type: TypeString},
			{Name: "name", Type: TypeString},
		})
		if !s.HasColumn("id") {
			t.Error("should have column 'id'")
		}
		if s.HasColumn("nope") {
			t.Error("should not have column 'nope'")
		}
	})

	t.Run("validate_required_field", func(t *testing.T) {
		s := New("test", []Column{
			{Name: "id", Type: TypeString, Required: true},
			{Name: "opt", Type: TypeString},
		})

		errs := s.Validate(map[string]string{"id": "REQ-001", "opt": ""})
		if len(errs) != 0 {
			t.Errorf("valid row should have no errors, got: %v", errs)
		}

		errs = s.Validate(map[string]string{"opt": "value"})
		if len(errs) != 1 || !strings.Contains(errs[0].Error(), "required") {
			t.Errorf("missing required should error, got: %v", errs)
		}
	})

	t.Run("validate_int_type", func(t *testing.T) {
		s := New("test", []Column{{Name: "count", Type: TypeInt}})

		errs := s.Validate(map[string]string{"count": "42"})
		if len(errs) != 0 {
			t.Errorf("valid int should pass, got: %v", errs)
		}

		errs = s.Validate(map[string]string{"count": "abc"})
		if len(errs) != 1 || !strings.Contains(errs[0].Error(), "expected int") {
			t.Errorf("invalid int should fail, got: %v", errs)
		}
	})

	t.Run("validate_float_type", func(t *testing.T) {
		s := New("test", []Column{{Name: "score", Type: TypeFloat}})

		errs := s.Validate(map[string]string{"score": "3.14"})
		if len(errs) != 0 {
			t.Errorf("valid float should pass, got: %v", errs)
		}

		errs = s.Validate(map[string]string{"score": "not-a-number"})
		if len(errs) != 1 {
			t.Errorf("invalid float should fail, got: %v", errs)
		}
	})

	t.Run("validate_bool_type", func(t *testing.T) {
		s := New("test", []Column{{Name: "active", Type: TypeBool}})

		for _, valid := range []string{"true", "false", "1", "0", "True", "FALSE"} {
			errs := s.Validate(map[string]string{"active": valid})
			if len(errs) != 0 {
				t.Errorf("bool %q should pass, got: %v", valid, errs)
			}
		}

		errs := s.Validate(map[string]string{"active": "yes"})
		if len(errs) != 1 {
			t.Errorf("invalid bool should fail, got: %v", errs)
		}
	})

	t.Run("validate_date_type", func(t *testing.T) {
		s := New("test", []Column{{Name: "d", Type: TypeDate}})

		errs := s.Validate(map[string]string{"d": "2026-05-08"})
		if len(errs) != 0 {
			t.Errorf("valid date should pass, got: %v", errs)
		}

		errs = s.Validate(map[string]string{"d": "05/08/2026"})
		if len(errs) != 1 {
			t.Errorf("invalid date format should fail, got: %v", errs)
		}
	})

	t.Run("validate_enum_type", func(t *testing.T) {
		s := New("test", []Column{{Name: "status", Type: TypeEnum, EnumValues: []string{"OPEN", "CLOSED"}}})

		errs := s.Validate(map[string]string{"status": "OPEN"})
		if len(errs) != 0 {
			t.Errorf("valid enum should pass, got: %v", errs)
		}

		errs = s.Validate(map[string]string{"status": "open"}) // case insensitive
		if len(errs) != 0 {
			t.Errorf("case-insensitive enum should pass, got: %v", errs)
		}

		errs = s.Validate(map[string]string{"status": "INVALID"})
		if len(errs) != 1 {
			t.Errorf("invalid enum should fail, got: %v", errs)
		}
	})

	t.Run("validate_set_type", func(t *testing.T) {
		s := New("test", []Column{{Name: "tags", Type: TypeSet, EnumValues: []string{"A", "B", "C"}}})

		errs := s.Validate(map[string]string{"tags": "A|B"})
		if len(errs) != 0 {
			t.Errorf("valid set should pass, got: %v", errs)
		}

		errs = s.Validate(map[string]string{"tags": "A|Z"})
		if len(errs) != 1 || !strings.Contains(errs[0].Error(), "not in") {
			t.Errorf("invalid set element should fail, got: %v", errs)
		}
	})

	t.Run("validate_header", func(t *testing.T) {
		s := New("test", []Column{
			{Name: "id", Type: TypeString, Required: true},
			{Name: "name", Type: TypeString},
			{Name: "status", Type: TypeEnum},
		})

		missing, extra := s.ValidateHeader([]string{"id", "name", "status"})
		if len(missing) != 0 || len(extra) != 0 {
			t.Errorf("exact match should have no missing/extra, got missing=%v extra=%v", missing, extra)
		}

		missing, extra = s.ValidateHeader([]string{"id", "custom_field"})
		if len(missing) != 2 { // name and status missing
			t.Errorf("expected 2 missing, got %v", missing)
		}
		if len(extra) != 1 || extra[0] != "custom_field" {
			t.Errorf("expected 1 extra (custom_field), got %v", extra)
		}
	})

	t.Run("extend_schema", func(t *testing.T) {
		base := New("core", []Column{
			{Name: "id", Type: TypeString, Required: true},
		})

		extended := base.Extend("extended", []Column{
			{Name: "custom", Type: TypeString},
		})

		if extended.Name != "extended" {
			t.Errorf("extended name = %q, want extended", extended.Name)
		}
		if len(extended.Columns) != 2 {
			t.Errorf("extended should have 2 columns, got %d", len(extended.Columns))
		}
		if !extended.HasColumn("custom") {
			t.Error("extended should have 'custom' column")
		}
		// Base should be unchanged
		if len(base.Columns) != 1 {
			t.Error("base should not be modified by extend")
		}
	})

	t.Run("empty_values_skip_validation", func(t *testing.T) {
		s := New("test", []Column{
			{Name: "count", Type: TypeInt},
			{Name: "score", Type: TypeFloat},
			{Name: "d", Type: TypeDate},
		})

		errs := s.Validate(map[string]string{"count": "", "score": "", "d": ""})
		if len(errs) != 0 {
			t.Errorf("empty non-required fields should pass, got: %v", errs)
		}
	})

	t.Run("column_type_string", func(t *testing.T) {
		types := []struct {
			t    ColumnType
			want string
		}{
			{TypeString, "string"},
			{TypeInt, "int"},
			{TypeFloat, "float"},
			{TypeBool, "bool"},
			{TypeDate, "date"},
			{TypeEnum, "enum"},
			{TypeSet, "set"},
		}
		for _, tt := range types {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("ColumnType(%d).String() = %q, want %q", tt.t, got, tt.want)
			}
		}
	})
}

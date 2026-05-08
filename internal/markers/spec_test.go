package markers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestMarkerSchemaValidation(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-007")

	t.Run("valid_full_marker", func(t *testing.T) {
		m := &Marker{
			ReqID:     "REQ-AUTH-001",
			TestName:  "TestLogin",
			TestFile:  "tests/auth_test.go",
			Scope:     "integration",
			Technique: "nominal",
			Env:       "simulation",
			Line:      42,
		}
		if errs := m.Validate(); len(errs) > 0 {
			t.Errorf("valid marker should have no errors, got: %v", errs)
		}
	})

	t.Run("valid_minimal_marker", func(t *testing.T) {
		m := &Marker{
			ReqID:    "REQ-GO-001",
			TestName: "TestBuild",
			TestFile: "cmd/build_test.go",
		}
		if errs := m.Validate(); len(errs) > 0 {
			t.Errorf("minimal marker should have no errors, got: %v", errs)
		}
	})

	t.Run("missing_req_id", func(t *testing.T) {
		m := &Marker{TestName: "TestFoo", TestFile: "foo_test.go"}
		errs := m.Validate()
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
		}
		if !strings.Contains(errs[0].Error(), "req_id") {
			t.Errorf("error should mention req_id: %v", errs[0])
		}
	})

	t.Run("invalid_req_id_format", func(t *testing.T) {
		m := &Marker{ReqID: "BAD-FORMAT", TestName: "TestFoo", TestFile: "foo_test.go"}
		errs := m.Validate()
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
		}
		if !strings.Contains(errs[0].Error(), "invalid req_id") {
			t.Errorf("error should mention invalid req_id: %v", errs[0])
		}
	})

	t.Run("missing_test_name", func(t *testing.T) {
		m := &Marker{ReqID: "REQ-GO-001", TestFile: "foo_test.go"}
		errs := m.Validate()
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
		}
		if !strings.Contains(errs[0].Error(), "test_name") {
			t.Errorf("error should mention test_name: %v", errs[0])
		}
	})

	t.Run("missing_test_file", func(t *testing.T) {
		m := &Marker{ReqID: "REQ-GO-001", TestName: "TestFoo"}
		errs := m.Validate()
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
		}
	})

	t.Run("invalid_scope", func(t *testing.T) {
		m := &Marker{ReqID: "REQ-GO-001", TestName: "TestFoo", TestFile: "foo_test.go", Scope: "invalid"}
		errs := m.Validate()
		if len(errs) != 1 || !strings.Contains(errs[0].Error(), "scope") {
			t.Errorf("expected scope error, got: %v", errs)
		}
	})

	t.Run("invalid_technique", func(t *testing.T) {
		m := &Marker{ReqID: "REQ-GO-001", TestName: "TestFoo", TestFile: "foo_test.go", Technique: "invalid"}
		errs := m.Validate()
		if len(errs) != 1 || !strings.Contains(errs[0].Error(), "technique") {
			t.Errorf("expected technique error, got: %v", errs)
		}
	})

	t.Run("invalid_env", func(t *testing.T) {
		m := &Marker{ReqID: "REQ-GO-001", TestName: "TestFoo", TestFile: "foo_test.go", Env: "invalid"}
		errs := m.Validate()
		if len(errs) != 1 || !strings.Contains(errs[0].Error(), "env") {
			t.Errorf("expected env error, got: %v", errs)
		}
	})

	t.Run("all_valid_scopes", func(t *testing.T) {
		for _, scope := range ValidScopes {
			m := &Marker{ReqID: "REQ-GO-001", TestName: "T", TestFile: "t.go", Scope: scope}
			if errs := m.Validate(); len(errs) > 0 {
				t.Errorf("scope %q should be valid, got: %v", scope, errs)
			}
		}
	})

	t.Run("all_valid_techniques", func(t *testing.T) {
		for _, tech := range ValidTechniques {
			m := &Marker{ReqID: "REQ-GO-001", TestName: "T", TestFile: "t.go", Technique: tech}
			if errs := m.Validate(); len(errs) > 0 {
				t.Errorf("technique %q should be valid, got: %v", tech, errs)
			}
		}
	})

	t.Run("all_valid_envs", func(t *testing.T) {
		for _, env := range ValidEnvs {
			m := &Marker{ReqID: "REQ-GO-001", TestName: "T", TestFile: "t.go", Env: env}
			if errs := m.Validate(); len(errs) > 0 {
				t.Errorf("env %q should be valid, got: %v", env, errs)
			}
		}
	})

	t.Run("json_roundtrip", func(t *testing.T) {
		m := &Marker{
			ReqID:     "REQ-AUTH-001",
			TestName:  "TestLogin",
			TestFile:  "tests/auth_test.go",
			Scope:     "unit",
			Technique: "nominal",
			Env:       "simulation",
			Line:      10,
		}

		data, err := json.Marshal(m)
		if err != nil {
			t.Fatalf("marshal failed: %v", err)
		}

		var m2 Marker
		if err := json.Unmarshal(data, &m2); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		if m2.ReqID != m.ReqID || m2.TestName != m.TestName || m2.TestFile != m.TestFile {
			t.Errorf("roundtrip mismatch: %+v vs %+v", m, m2)
		}
	})

	t.Run("json_schema_file_exists", func(t *testing.T) {
		wd, _ := os.Getwd()
		projectRoot := filepath.Dir(filepath.Dir(wd))
		if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
			projectRoot = filepath.Dir(wd)
		}

		schemaPath := filepath.Join(projectRoot, "docs", "marker-schema.json")
		data, err := os.ReadFile(schemaPath)
		if err != nil {
			t.Fatalf("marker-schema.json not found: %v", err)
		}

		var schema map[string]interface{}
		if err := json.Unmarshal(data, &schema); err != nil {
			t.Fatalf("marker-schema.json is not valid JSON: %v", err)
		}

		// Verify required fields match our spec
		required, ok := schema["required"].([]interface{})
		if !ok {
			t.Fatal("schema should have required array")
		}
		requiredSet := map[string]bool{}
		for _, r := range required {
			requiredSet[r.(string)] = true
		}
		for _, field := range []string{"req_id", "test_name", "test_file"} {
			if !requiredSet[field] {
				t.Errorf("schema should require field %q", field)
			}
		}

		// Verify properties match our spec
		props, ok := schema["properties"].(map[string]interface{})
		if !ok {
			t.Fatal("schema should have properties object")
		}
		for _, field := range []string{"req_id", "test_name", "test_file", "scope", "technique", "env", "line"} {
			if _, ok := props[field]; !ok {
				t.Errorf("schema should define property %q", field)
			}
		}
	})

	t.Run("multiple_errors_reported", func(t *testing.T) {
		m := &Marker{} // all required fields missing
		errs := m.Validate()
		if len(errs) < 3 {
			t.Errorf("expected at least 3 errors for empty marker, got %d: %v", len(errs), errs)
		}
	})
}

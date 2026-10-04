// Package docmodel validates Requirement Document Model fixtures (REQ-DATA-001a).
// Spike only: does not change production CSV or CLI defaults.
package docmodel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

var (
	schemaOnce sync.Once
	schema     *jsonschema.Schema
	schemaErr  error
)

// SchemaPath returns the path to requirement-document-v0.schema.json.
func SchemaPath() string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, "docs", "schemas", "requirement-document-v0.schema.json")
}

func loadSchema() (*jsonschema.Schema, error) {
	schemaOnce.Do(func() {
		path := SchemaPath()
		data, err := os.ReadFile(path)
		if err != nil {
			schemaErr = err
			return
		}
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
		if err != nil {
			schemaErr = err
			return
		}
		c := jsonschema.NewCompiler()
		if err := c.AddResource("requirement-document-v0.schema.json", doc); err != nil {
			schemaErr = err
			return
		}
		schema, schemaErr = c.Compile("requirement-document-v0.schema.json")
	})
	return schema, schemaErr
}

// ValidateDocument validates a Requirement Document Model JSON value.
func ValidateDocument(doc any) error {
	s, err := loadSchema()
	if err != nil {
		return fmt.Errorf("load schema: %w", err)
	}
	if err := s.Validate(doc); err != nil {
		return err
	}
	return nil
}

// ValidateJSON validates raw JSON bytes against the schema.
func ValidateJSON(raw []byte) error {
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}
	return ValidateDocument(doc)
}

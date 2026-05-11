package schema

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// CustomSchemaConfig represents a .rtmx/schema.yaml file that defines
// custom columns to extend the base schema.
type CustomSchemaConfig struct {
	Name    string             `yaml:"name"`
	Extends string             `yaml:"extends"` // base schema name (default: core)
	Columns []CustomColumnDef  `yaml:"columns"`
}

// CustomColumnDef defines a custom column in schema.yaml.
type CustomColumnDef struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"` // string, int, float, bool, date, enum, set
	Required    bool     `yaml:"required"`
	Description string   `yaml:"description"`
	Values      []string `yaml:"values"` // for enum/set types
}

// LoadCustomSchema reads a schema.yaml file and returns a Schema
// that extends the named base schema with custom columns.
func LoadCustomSchema(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema config: %w", err)
	}

	var cfg CustomSchemaConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse schema config: %w", err)
	}

	if cfg.Name == "" {
		cfg.Name = "custom"
	}
	if cfg.Extends == "" {
		cfg.Extends = "core"
	}

	base := Get(cfg.Extends)
	if base == nil {
		return nil, fmt.Errorf("base schema %q not found", cfg.Extends)
	}

	columns := make([]Column, 0, len(cfg.Columns))
	for _, cd := range cfg.Columns {
		col := Column{
			Name:        cd.Name,
			Type:        parseColumnType(cd.Type),
			Required:    cd.Required,
			Description: cd.Description,
			EnumValues:  cd.Values,
		}
		columns = append(columns, col)
	}

	return base.Extend(cfg.Name, columns), nil
}

func parseColumnType(s string) ColumnType {
	switch strings.ToLower(s) {
	case "int", "integer":
		return TypeInt
	case "float", "number":
		return TypeFloat
	case "bool", "boolean":
		return TypeBool
	case "date":
		return TypeDate
	case "enum":
		return TypeEnum
	case "set":
		return TypeSet
	default:
		return TypeString
	}
}

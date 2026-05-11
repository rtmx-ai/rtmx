package schema

import (
	"fmt"
	"sort"
	"sync"
)

// registry is the global schema registry.
var registry = &Registry{
	schemas: make(map[string]*Schema),
}

// Registry holds named schemas and provides lookup.
type Registry struct {
	mu      sync.RWMutex
	schemas map[string]*Schema
}

// Register adds a schema to the global registry.
// Panics if a schema with the same name is already registered.
func Register(s *Schema) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, exists := registry.schemas[s.Name]; exists {
		panic(fmt.Sprintf("schema %q already registered", s.Name))
	}
	registry.schemas[s.Name] = s
}

// Get retrieves a schema by name from the global registry.
// Returns nil if not found.
func Get(name string) *Schema {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	return registry.schemas[name]
}

// Names returns all registered schema names, sorted.
func Names() []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	names := make([]string, 0, len(registry.schemas))
	for name := range registry.schemas {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ForConfig returns the schema for the given config schema name.
// Falls back to "core" if the name is empty. Returns an error if
// the schema is not registered.
func ForConfig(schemaName string) (*Schema, error) {
	if schemaName == "" {
		schemaName = "core"
	}
	s := Get(schemaName)
	if s == nil {
		return nil, fmt.Errorf("unknown schema %q (registered: %v)", schemaName, Names())
	}
	return s, nil
}

func init() {
	// Register built-in schemas
	Register(CoreSchema)
}

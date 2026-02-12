// Package rtmx provides requirement traceability markers for Go tests.
//
// RTMX enables requirement-driven development by linking test functions
// to specific requirements in your RTM database.
//
// Basic usage:
//
//	func TestLogin(t *testing.T) {
//	    rtmx.Req(t, "REQ-AUTH-001")
//	    // test implementation
//	}
//
// With options:
//
//	func TestLogin(t *testing.T) {
//	    rtmx.Req(t, "REQ-AUTH-001",
//	        rtmx.Scope("integration"),
//	        rtmx.Technique("nominal"),
//	        rtmx.Env("simulation"),
//	    )
//	    // test implementation
//	}
//
// Table-driven tests with struct tags:
//
//	func TestCalculations(t *testing.T) {
//	    tests := []struct {
//	        name     string
//	        rtmx     string `rtmx:"REQ-MATH-001"`
//	        input    int
//	        expected int
//	    }{
//	        {"positive", "", 5, 25},
//	    }
//	    for _, tt := range tests {
//	        t.Run(tt.name, func(t *testing.T) {
//	            rtmx.FromTag(t, tt)
//	            // test implementation
//	        })
//	    }
//	}
package rtmx

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sync"
	"testing"
	"time"
)

// reqIDPattern validates requirement ID format.
var reqIDPattern = regexp.MustCompile(`^REQ-[A-Z]+-[0-9]+$`)

// marker holds the requirement marker data for a test.
type marker struct {
	ReqID     string `json:"req_id"`
	Scope     string `json:"scope,omitempty"`
	Technique string `json:"technique,omitempty"`
	Env       string `json:"env,omitempty"`
	TestName  string `json:"test_name"`
	TestFile  string `json:"test_file"`
	Line      int    `json:"line"`
}

// testResult holds the result of a test with requirement marker.
type testResult struct {
	Marker    marker    `json:"marker"`
	Passed    bool      `json:"passed"`
	Duration  float64   `json:"duration_ms"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// registry stores all registered markers and results.
type registry struct {
	mu      sync.Mutex
	markers map[string]*marker
	results []testResult
}

var globalRegistry = &registry{
	markers: make(map[string]*marker),
	results: []testResult{},
}

// Option configures a requirement marker.
type Option func(*marker)

// Scope sets the test scope (unit, integration, system, e2e).
func Scope(s string) Option {
	return func(m *marker) {
		m.Scope = s
	}
}

// Technique sets the test technique (nominal, boundary, error, stress).
func Technique(t string) Option {
	return func(m *marker) {
		m.Technique = t
	}
}

// Env sets the test environment (simulation, hil, field).
func Env(e string) Option {
	return func(m *marker) {
		m.Env = e
	}
}

// Req registers a requirement marker for the current test.
//
// The reqID must match the pattern REQ-[A-Z]+-[0-9]+.
// If the pattern doesn't match, the test fails immediately.
//
// Example:
//
//	func TestFeature(t *testing.T) {
//	    rtmx.Req(t, "REQ-FEAT-001",
//	        rtmx.Scope("unit"),
//	        rtmx.Technique("nominal"),
//	    )
//	    // test implementation
//	}
func Req(t testing.TB, reqID string, opts ...Option) {
	t.Helper()

	if !reqIDPattern.MatchString(reqID) {
		t.Fatalf("rtmx: invalid requirement ID format: %s (expected REQ-[A-Z]+-[0-9]+)", reqID)
	}

	// Get caller info
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	m := &marker{
		ReqID:    reqID,
		TestName: t.Name(),
		TestFile: filepath.Base(file),
		Line:     line,
	}

	for _, opt := range opts {
		opt(m)
	}

	register(t, m)
}

// FromTag extracts requirement from struct tag and registers it.
//
// Use this for table-driven tests where the requirement is defined
// in a struct tag on the test case.
//
// Example:
//
//	tests := []struct {
//	    name string
//	    rtmx string `rtmx:"REQ-MATH-001"`
//	    input int
//	}{
//	    {"case1", "", 5},
//	}
//	for _, tt := range tests {
//	    t.Run(tt.name, func(t *testing.T) {
//	        rtmx.FromTag(t, tt)
//	        // test
//	    })
//	}
func FromTag(t testing.TB, testCase interface{}) {
	t.Helper()

	reqID := extractTagValue(testCase, "rtmx")
	if reqID == "" {
		return // No rtmx tag found, silently skip
	}

	Req(t, reqID)
}

// register stores the marker in the global registry.
func register(t testing.TB, m *marker) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	key := fmt.Sprintf("%s:%s", m.TestFile, m.TestName)
	globalRegistry.markers[key] = m

	// Register cleanup to record test result
	t.Cleanup(func() {
		recordResult(t, m)
	})
}

// recordResult records the test result when the test completes.
func recordResult(t testing.TB, m *marker) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	result := testResult{
		Marker:    *m,
		Passed:    !t.Failed(),
		Timestamp: time.Now(),
	}

	globalRegistry.results = append(globalRegistry.results, result)
}

// extractTagValue extracts a value from a struct tag.
func extractTagValue(v interface{}, tagName string) string {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return ""
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if tag := field.Tag.Get(tagName); tag != "" {
			return tag
		}
	}

	return ""
}

// Results returns all recorded test results.
func Results() []testResult {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	results := make([]testResult, len(globalRegistry.results))
	copy(results, globalRegistry.results)
	return results
}

// WriteResultsJSON writes test results to a JSON file.
//
// Call this in TestMain after m.Run() completes:
//
//	func TestMain(m *testing.M) {
//	    code := m.Run()
//	    rtmx.WriteResultsJSON("rtmx-results.json")
//	    os.Exit(code)
//	}
func WriteResultsJSON(path string) error {
	results := Results()
	if len(results) == 0 {
		return nil
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}

	return nil
}

// ClearRegistry clears all registered markers and results.
// Useful for testing the rtmx package itself.
func ClearRegistry() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	globalRegistry.markers = make(map[string]*marker)
	globalRegistry.results = []testResult{}
}

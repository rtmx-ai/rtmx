// Package testutil provides test utilities and fixtures for RTMX Go CLI testing.
package testutil

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var updateGolden = flag.Bool("update", false, "update golden files")

// Update returns true if golden files should be updated.
// Use with: go test -update
func Update() bool {
	return *updateGolden
}

// Golden compares actual output against a golden file.
// If the -update flag is set, it writes the actual output to the golden file.
// The golden file is stored in the testdata directory relative to the test file.
//
// Usage:
//
//	func TestOutput(t *testing.T) {
//	    actual := generateOutput()
//	    testutil.Golden(t, "expected_output", actual)
//	}
//
// To update golden files:
//
//	go test -update ./...
//
// Golden files are stored as testdata/<name>.golden
func Golden(t *testing.T, name string, actual []byte) {
	t.Helper()

	goldenPath := filepath.Join("testdata", name+".golden")

	if Update() {
		// Create testdata directory if it doesn't exist
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("Failed to create testdata directory: %v", err)
		}

		// Write actual output to golden file
		if err := os.WriteFile(goldenPath, actual, 0644); err != nil {
			t.Fatalf("Failed to write golden file %s: %v", goldenPath, err)
		}
		t.Logf("Updated golden file: %s", goldenPath)
		return
	}

	// Read expected output from golden file
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Golden file %s does not exist. Run with -update to create it.", goldenPath)
		}
		t.Fatalf("Failed to read golden file %s: %v", goldenPath, err)
	}

	// Compare actual with expected
	if string(actual) != string(expected) {
		t.Errorf("Output does not match golden file %s.\n"+
			"To update the golden file, run: go test -update ./...\n\n"+
			"Got:\n%s\n\nWant:\n%s",
			goldenPath, string(actual), string(expected))
	}
}

// GoldenString is a convenience wrapper for Golden that accepts a string.
func GoldenString(t *testing.T, name string, actual string) {
	t.Helper()
	Golden(t, name, []byte(actual))
}

// StripANSI removes ANSI escape codes from a byte slice.
// Useful for creating platform-independent golden files.
func StripANSI(data []byte) []byte {
	return []byte(StripANSIString(string(data)))
}

// StripANSIString removes ANSI escape codes from a string.
func StripANSIString(s string) string {
	var result []byte
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z') {
				inEscape = false
			}
			continue
		}
		result = append(result, s[i])
	}
	return string(result)
}

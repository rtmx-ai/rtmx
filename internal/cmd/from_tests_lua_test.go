package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractLuaMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-027")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker -- rtmx:req REQ-ID with function",
			content: `-- rtmx:req REQ-GAME-001
function test_player_move()
    assert(true)
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-GAME-001", TestFunction: "test_player_move"},
			},
		},
		{
			name: "comment marker with it block (busted)",
			content: `describe("player", function()
    -- rtmx:req REQ-GAME-002
    it("moves correctly", function()
        assert.is_true(true)
    end)
end)
`,
			expected: []TestRequirement{
				{ReqID: "REQ-GAME-002", TestFunction: "moves correctly"},
			},
		},
		{
			name: "local function",
			content: `-- rtmx:req REQ-GAME-003
local function test_collision()
    assert(true)
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-GAME-003", TestFunction: "test_collision"},
			},
		},
		{
			name: "multiple markers",
			content: `-- rtmx:req REQ-GAME-001
function test_move()
    assert(true)
end

-- rtmx:req REQ-GAME-002
function test_attack()
    assert(true)
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-GAME-001", TestFunction: "test_move"},
				{ReqID: "REQ-GAME-002", TestFunction: "test_attack"},
			},
		},
		{
			name: "no markers",
			content: `function test_no_marker()
    assert(true)
end
`,
			expected: nil,
		},
		{
			name: "empty file",
			content: ``,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rtmx-lua-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "game_test.lua")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractLuaMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractLuaMarkersFromFile failed: %v", err)
			}

			if len(markers) != len(tt.expected) {
				t.Fatalf("Expected %d markers, got %d", len(tt.expected), len(markers))
			}

			for i, exp := range tt.expected {
				got := markers[i]
				if got.ReqID != exp.ReqID {
					t.Errorf("Marker %d: expected ReqID %q, got %q", i, exp.ReqID, got.ReqID)
				}
				if got.TestFunction != exp.TestFunction {
					t.Errorf("Marker %d: expected TestFunction %q, got %q", i, exp.TestFunction, got.TestFunction)
				}
				if got.TestFile != testFile {
					t.Errorf("Marker %d: expected TestFile %q, got %q", i, testFile, got.TestFile)
				}
				if got.LineNumber == 0 {
					t.Errorf("Marker %d: expected non-zero line number", i)
				}
			}
		})
	}
}

func TestExtractLuaMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-027")

	_, err := extractLuaMarkersFromFile("/nonexistent/game_test.lua")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsLuaTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-027")

	tests := []struct {
		path     string
		expected bool
	}{
		{"game_test.lua", true},
		{"test_game.lua", true},
		{"luacheck_spec.lua", true},
		{"config_spec.lua", true},
		{"spec/feature_spec.lua", true},
		{"game.lua", false},
		{"game_test.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isLuaTestFile(tt.path); got != tt.expected {
				t.Errorf("isLuaTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

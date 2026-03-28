package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractVerilogMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-015")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker // rtmx:req REQ-ID with task",
			content: `module test_alu;
    // rtmx:req REQ-ALU-001
    task test_add;
        begin
            $display("testing add");
        end
    endtask
endmodule
`,
			expected: []TestRequirement{
				{ReqID: "REQ-ALU-001", TestFunction: "test_alu.test_add"},
			},
		},
		{
			name: "multiple markers",
			content: `module test_cpu;
    // rtmx:req REQ-CPU-001
    task test_fetch;
        begin
            $display("testing fetch");
        end
    endtask

    // rtmx:req REQ-CPU-002
    task test_decode;
        begin
            $display("testing decode");
        end
    endtask
endmodule
`,
			expected: []TestRequirement{
				{ReqID: "REQ-CPU-001", TestFunction: "test_cpu.test_fetch"},
				{ReqID: "REQ-CPU-002", TestFunction: "test_cpu.test_decode"},
			},
		},
		{
			name: "no markers",
			content: `module test_alu;
    task test_add;
        begin
            $display("testing add");
        end
    endtask
endmodule
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
			tmpDir, err := os.MkdirTemp("", "rtmx-verilog-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "alu_test.sv")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractVerilogMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractVerilogMarkersFromFile failed: %v", err)
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

func TestExtractVerilogMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-015")

	_, err := extractVerilogMarkersFromFile("/nonexistent/alu_test.sv")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsVerilogTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-015")

	tests := []struct {
		path     string
		expected bool
	}{
		{"alu_test.sv", true},
		{"alu_test.v", true},
		{"alu_tb.sv", true},
		{"alu_tb.v", true},
		{"alu.sv", false},
		{"alu.v", false},
		{"alu_test.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isVerilogTestFile(tt.path); got != tt.expected {
				t.Errorf("isVerilogTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

package docmodel

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func fixturesDir(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, "docs", "schemas", "fixtures")
}

func TestValidFixtureThreeRequirements(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-001a")
	raw, err := os.ReadFile(filepath.Join(fixturesDir(t), "requirement-document-v0.valid.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateJSON(raw); err != nil {
		t.Fatalf("expected valid fixture: %v", err)
	}
}

func TestInvalidDocumentsReportJSONPointerPaths(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-001a")

	cases := []struct {
		file    string
		wantSub string
	}{
		{
			file:    "requirement-document-v0.invalid-req-id.json",
			wantSub: "/requirements/0/req_id",
		},
		{
			file:    "requirement-document-v0.invalid-ac.json",
			wantSub: "/requirements/0/acs/0",
		},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(fixturesDir(t), tc.file))
			if err != nil {
				t.Fatal(err)
			}
			err = ValidateJSON(raw)
			if err == nil {
				t.Fatal("expected validation error")
			}
			msg := err.Error()
			if !strings.Contains(msg, tc.wantSub) {
				t.Fatalf("error %q missing pointer %q", msg, tc.wantSub)
			}
		})
	}
}

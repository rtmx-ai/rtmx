package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
)

// TestCIValidate validates that a reusable validation workflow exists
// for other RTMX-enabled repos to call.
// REQ-CI-006: Reusable validation workflow
func TestCIValidate(t *testing.T) {
	rtmx.Req(t, "REQ-CI-006")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	wfPath := filepath.Join(projectRoot, ".github", "workflows", "rtmx-validate.yml")
	content, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatalf("Failed to read rtmx-validate.yml: %v", err)
	}
	wf := string(content)

	// AC1: Callable via workflow_call
	t.Run("workflow_call_trigger", func(t *testing.T) {
		if !strings.Contains(wf, "workflow_call:") {
			t.Fatal("Workflow must be callable via workflow_call")
		}
	})

	// AC2: Configurable RTM database path
	t.Run("configurable_csv_path", func(t *testing.T) {
		if !strings.Contains(wf, "rtm-csv-path:") {
			t.Error("Workflow must accept rtm-csv-path input")
		}
		if !strings.Contains(wf, "default:") {
			t.Error("rtm-csv-path must have a default value")
		}
	})

	// AC3: Outputs health status
	t.Run("outputs_health_status", func(t *testing.T) {
		if !strings.Contains(wf, "outputs:") {
			t.Fatal("Workflow must define outputs")
		}
		if !strings.Contains(wf, "status:") {
			t.Error("Workflow must output health status")
		}
	})

	// AC4: Uploads health report as artifact
	t.Run("uploads_artifact", func(t *testing.T) {
		if !strings.Contains(wf, "actions/upload-artifact") {
			t.Error("Workflow must upload health report artifact")
		}
	})

	// AC5: Fails on unhealthy status
	t.Run("fails_on_unhealthy", func(t *testing.T) {
		if !strings.Contains(wf, "unhealthy") {
			t.Error("Workflow must fail on unhealthy status")
		}
	})
}

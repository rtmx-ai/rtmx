package sync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
)

const csvHeader = "req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id\n"

// createTempProject creates a temp directory with a .rtmx/database.csv
func createTempProject(t *testing.T, csvRows string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "rtmx-shadow-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	rtmxDir := filepath.Join(dir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		t.Fatalf("Failed to create .rtmx dir: %v", err)
	}

	dbPath := filepath.Join(rtmxDir, "database.csv")
	if err := os.WriteFile(dbPath, []byte(csvHeader+csvRows), 0644); err != nil {
		t.Fatalf("Failed to write database: %v", err)
	}
	return dir
}

func TestShadowRequirements(t *testing.T) {
	rtmx.Req(t, "REQ-GO-035")

	// Set up a "remote" project with two requirements
	remoteDir := createTempProject(t,
		"REQ-AUTH-001,AUTH,Login,User authentication,Login works,auth_test.go,TestAuth,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n"+
			"REQ-AUTH-002,AUTH,Token,Token refresh,Tokens refresh,token_test.go,TestToken,Unit Test,MISSING,HIGH,1,,,,REQ-AUTH-001,,,,,,\n")

	// Set up a "local" project that depends on the remote
	localDir := createTempProject(t,
		"REQ-GO-100,CLI,Feature,Local feature,Feature works,feat_test.go,TestFeat,Unit Test,MISSING,HIGH,1,,,sync:upstream:REQ-AUTH-001,,,,,,\n"+
			"REQ-GO-101,CLI,Feature,Another feature,Works,feat2_test.go,TestFeat2,Unit Test,MISSING,HIGH,1,,,sync:upstream:REQ-AUTH-002,,,,,,\n"+
			"REQ-GO-102,CLI,Feature,Bad ref,Works,feat3_test.go,TestFeat3,Unit Test,MISSING,HIGH,1,,,sync:nonexistent:REQ-999,,,,,,\n")

	remotes := map[string]config.SyncRemote{
		"upstream": {
			Repo:     "rtmx-ai/rtmx",
			Database: ".rtmx/database.csv",
			Path:     remoteDir,
		},
	}
	_ = localDir // used contextually

	t.Run("ParseRef_valid", func(t *testing.T) {
		alias, reqID, err := ParseRef("sync:upstream:REQ-AUTH-001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if alias != "upstream" {
			t.Errorf("expected alias 'upstream', got %q", alias)
		}
		if reqID != "REQ-AUTH-001" {
			t.Errorf("expected reqID 'REQ-AUTH-001', got %q", reqID)
		}
	})

	t.Run("ParseRef_invalid_no_prefix", func(t *testing.T) {
		_, _, err := ParseRef("REQ-AUTH-001")
		if err == nil {
			t.Fatal("expected error for non-sync ref")
		}
	})

	t.Run("ParseRef_invalid_missing_parts", func(t *testing.T) {
		_, _, err := ParseRef("sync:upstream")
		if err == nil {
			t.Fatal("expected error for missing req ID")
		}
	})

	t.Run("ParseRef_invalid_empty_alias", func(t *testing.T) {
		_, _, err := ParseRef("sync::REQ-001")
		if err == nil {
			t.Fatal("expected error for empty alias")
		}
	})

	t.Run("IsResolvable", func(t *testing.T) {
		if !IsResolvable("sync:upstream:REQ-001") {
			t.Error("sync: prefix should be resolvable")
		}
		if IsResolvable("REQ-001") {
			t.Error("plain req ID should not be resolvable")
		}
		if IsResolvable("") {
			t.Error("empty string should not be resolvable")
		}
	})

	t.Run("Resolve_complete_shadow", func(t *testing.T) {
		resolver := NewShadowResolver(remotes)
		shadow, err := resolver.Resolve("sync:upstream:REQ-AUTH-001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if shadow.ReqID != "REQ-AUTH-001" {
			t.Errorf("expected ReqID 'REQ-AUTH-001', got %q", shadow.ReqID)
		}
		if shadow.RemoteAlias != "upstream" {
			t.Errorf("expected alias 'upstream', got %q", shadow.RemoteAlias)
		}
		if shadow.RemoteRepo != "rtmx-ai/rtmx" {
			t.Errorf("expected repo 'rtmx-ai/rtmx', got %q", shadow.RemoteRepo)
		}
		if !shadow.Status.IsComplete() {
			t.Errorf("expected COMPLETE status, got %v", shadow.Status)
		}
		if shadow.Description != "User authentication" {
			t.Errorf("expected description 'User authentication', got %q", shadow.Description)
		}
		if shadow.Visibility != "full" {
			t.Errorf("expected visibility 'full', got %q", shadow.Visibility)
		}
		if shadow.ResolvedAt.IsZero() {
			t.Error("ResolvedAt should be set")
		}
	})

	t.Run("Resolve_missing_shadow", func(t *testing.T) {
		resolver := NewShadowResolver(remotes)
		shadow, err := resolver.Resolve("sync:upstream:REQ-AUTH-002")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if shadow.Status.IsComplete() {
			t.Error("expected MISSING status for REQ-AUTH-002")
		}
	})

	t.Run("Resolve_caches_result", func(t *testing.T) {
		resolver := NewShadowResolver(remotes)
		s1, _ := resolver.Resolve("sync:upstream:REQ-AUTH-001")
		s2, _ := resolver.Resolve("sync:upstream:REQ-AUTH-001")
		if s1 != s2 {
			t.Error("expected same pointer from cache")
		}
	})

	t.Run("Resolve_unknown_alias", func(t *testing.T) {
		resolver := NewShadowResolver(remotes)
		_, err := resolver.Resolve("sync:nonexistent:REQ-001")
		if err == nil {
			t.Fatal("expected error for unknown alias")
		}
		if !contains(err.Error(), "unknown remote alias") {
			t.Errorf("expected 'unknown remote alias' error, got: %v", err)
		}
	})

	t.Run("Resolve_unknown_requirement", func(t *testing.T) {
		resolver := NewShadowResolver(remotes)
		_, err := resolver.Resolve("sync:upstream:REQ-NONEXISTENT")
		if err == nil {
			t.Fatal("expected error for unknown requirement")
		}
		if !contains(err.Error(), "not found") {
			t.Errorf("expected 'not found' error, got: %v", err)
		}
	})

	t.Run("Resolve_no_local_path", func(t *testing.T) {
		noPathRemotes := map[string]config.SyncRemote{
			"cloud": {Repo: "rtmx-ai/rtmx-cloud", Database: ".rtmx/database.csv"},
		}
		resolver := NewShadowResolver(noPathRemotes)
		_, err := resolver.Resolve("sync:cloud:REQ-001")
		if err == nil {
			t.Fatal("expected error for remote with no path")
		}
		if !contains(err.Error(), "no local path") {
			t.Errorf("expected 'no local path' error, got: %v", err)
		}
	})

	t.Run("Resolve_path_does_not_exist", func(t *testing.T) {
		badPathRemotes := map[string]config.SyncRemote{
			"gone": {Repo: "rtmx-ai/gone", Database: ".rtmx/database.csv", Path: "/nonexistent/path"},
		}
		resolver := NewShadowResolver(badPathRemotes)
		_, err := resolver.Resolve("sync:gone:REQ-001")
		if err == nil {
			t.Fatal("expected error for nonexistent path")
		}
	})

	t.Run("IsShadowBlocking_complete_not_blocking", func(t *testing.T) {
		resolver := NewShadowResolver(remotes)
		if resolver.IsShadowBlocking("sync:upstream:REQ-AUTH-001") {
			t.Error("COMPLETE shadow should not be blocking")
		}
	})

	t.Run("IsShadowBlocking_missing_is_blocking", func(t *testing.T) {
		resolver := NewShadowResolver(remotes)
		if !resolver.IsShadowBlocking("sync:upstream:REQ-AUTH-002") {
			t.Error("MISSING shadow should be blocking")
		}
	})

	t.Run("IsShadowBlocking_unresolvable_not_blocking", func(t *testing.T) {
		resolver := NewShadowResolver(remotes)
		if resolver.IsShadowBlocking("sync:nonexistent:REQ-999") {
			t.Error("unresolvable shadow should not block (permissive)")
		}
	})

	t.Run("ResolveAll_mixed_deps", func(t *testing.T) {
		localDB, err := loadDB(localDir)
		if err != nil {
			t.Fatalf("failed to load local db: %v", err)
		}

		resolver := NewShadowResolver(remotes)
		shadows := resolver.ResolveAll(localDB)
		warnings := resolver.Warnings()

		// Should resolve 2 shadows (AUTH-001 and AUTH-002) and 1 warning (nonexistent)
		if len(shadows) != 2 {
			t.Errorf("expected 2 resolved shadows, got %d", len(shadows))
		}
		if len(warnings) != 1 {
			t.Errorf("expected 1 warning, got %d", len(warnings))
		}
		if len(warnings) > 0 && !contains(warnings[0].Message, "unknown remote alias") {
			t.Errorf("expected unknown alias warning, got: %s", warnings[0].Message)
		}
	})

	t.Run("Warning_String", func(t *testing.T) {
		w := Warning{Ref: "sync:x:REQ-1", Message: "not found"}
		s := w.String()
		if !contains(s, "sync:x:REQ-1") || !contains(s, "not found") {
			t.Errorf("unexpected warning string: %s", s)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsImpl(s, substr)
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func loadDB(dir string) (*database.Database, error) {
	return database.Load(filepath.Join(dir, ".rtmx", "database.csv"))
}

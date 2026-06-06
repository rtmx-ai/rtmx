package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// setupGrantTestDir creates a temp project directory with config and a remote
// named "upstream" so that grant create can validate the grantee.
func setupGrantTestDir(t *testing.T) (tmpDir string, configPath string, cleanup func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "rtmx-grant-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create .rtmx dir: %v", err)
	}

	cfg := config.DefaultConfig()
	// Add a remote so grant create can validate the grantee
	if cfg.RTMX.Sync.Remotes == nil {
		cfg.RTMX.Sync.Remotes = make(map[string]config.SyncRemote)
	}
	cfg.RTMX.Sync.Remotes["upstream"] = config.SyncRemote{
		Repo:     "rtmx-ai/rtmx",
		Database: ".rtmx/database.csv",
	}
	cfg.RTMX.Sync.Remotes["partner"] = config.SyncRemote{
		Repo:     "partner-org/project",
		Database: ".rtmx/database.csv",
	}

	configPath = filepath.Join(rtmxDir, "config.yaml")
	if err := cfg.Save(configPath); err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("Failed to save config: %v", err)
	}

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("Failed to chdir: %v", err)
	}

	cleanup = func() {
		_ = os.Chdir(origDir)
		_ = os.RemoveAll(tmpDir)
	}

	return tmpDir, configPath, cleanup
}

func TestGrantCreate(t *testing.T) {
	rtmx.Req(t, "REQ-GO-079")

	_, configPath, cleanup := setupGrantTestDir(t)
	defer cleanup()

	t.Run("valid_grant_writes_to_config", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := grantCreateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		grantRole = "status_observer"
		grantCategories = nil
		grantIDs = nil
		grantExcludeCategories = nil
		grantExpiresAt = ""

		if err := cmd.RunE(cmd, []string{"upstream"}); err != nil {
			t.Fatalf("grant create failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Created grant") {
			t.Errorf("Expected 'Created grant' message, got: %s", out)
		}
		if !strings.Contains(out, "Grantee: upstream") {
			t.Errorf("Expected grantee in output, got: %s", out)
		}
		if !strings.Contains(out, "Role: status_observer") {
			t.Errorf("Expected role in output, got: %s", out)
		}
		if !strings.Contains(out, "visibility: shadow") {
			t.Errorf("Expected visibility in output, got: %s", out)
		}

		// Verify config was saved
		reloaded, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to reload config: %v", err)
		}
		if len(reloaded.RTMX.Sync.Grants) != 1 {
			t.Fatalf("Expected 1 grant, got %d", len(reloaded.RTMX.Sync.Grants))
		}
		g := reloaded.RTMX.Sync.Grants[0]
		if g.Grantee != "upstream" {
			t.Errorf("Expected grantee 'upstream', got %q", g.Grantee)
		}
		if g.Role != "status_observer" {
			t.Errorf("Expected role 'status_observer', got %q", g.Role)
		}
		if !strings.HasPrefix(g.ID, "grant-upstream-") {
			t.Errorf("Expected ID to start with 'grant-upstream-', got %q", g.ID)
		}
	})

	t.Run("invalid_role_returns_error", func(t *testing.T) {
		grantRole = "superuser"
		grantCategories = nil
		grantIDs = nil
		grantExcludeCategories = nil
		grantExpiresAt = ""

		err := grantCreateCmd.RunE(grantCreateCmd, []string{"partner"})
		if err == nil {
			t.Fatal("Expected error for invalid role")
		}
		if !strings.Contains(err.Error(), "invalid role") {
			t.Errorf("Expected 'invalid role' error, got: %v", err)
		}
	})

	t.Run("unknown_remote_returns_error", func(t *testing.T) {
		grantRole = "admin"
		grantCategories = nil
		grantIDs = nil
		grantExcludeCategories = nil
		grantExpiresAt = ""

		err := grantCreateCmd.RunE(grantCreateCmd, []string{"nonexistent"})
		if err == nil {
			t.Fatal("Expected error for unknown remote")
		}
		if !strings.Contains(err.Error(), "unknown remote") {
			t.Errorf("Expected 'unknown remote' error, got: %v", err)
		}
	})

	t.Run("expires_flag_stores_correctly", func(t *testing.T) {
		grantRole = "dependency_viewer"
		grantCategories = nil
		grantIDs = nil
		grantExcludeCategories = nil
		grantExpiresAt = "2027-12-31"

		buf := new(bytes.Buffer)
		cmd := grantCreateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		if err := cmd.RunE(cmd, []string{"partner"}); err != nil {
			t.Fatalf("grant create with --expires failed: %v", err)
		}

		reloaded, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to reload config: %v", err)
		}

		// Find the partner grant
		var found *config.SyncGrant
		for i, g := range reloaded.RTMX.Sync.Grants {
			if g.Grantee == "partner" {
				found = &reloaded.RTMX.Sync.Grants[i]
				break
			}
		}
		if found == nil {
			t.Fatal("Partner grant not found in config")
		}
		if found.Constraints.ExpiresAt != "2027-12-31" {
			t.Errorf("Expected ExpiresAt '2027-12-31', got %q", found.Constraints.ExpiresAt)
		}
	})

	t.Run("categories_constraint_stores_correctly", func(t *testing.T) {
		grantRole = "admin"
		grantCategories = []string{"AUTH", "API"}
		grantIDs = nil
		grantExcludeCategories = nil
		grantExpiresAt = ""

		buf := new(bytes.Buffer)
		cmd := grantCreateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		if err := cmd.RunE(cmd, []string{"partner"}); err != nil {
			t.Fatalf("grant create with --categories failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Categories:") {
			t.Errorf("Expected categories in output, got: %s", out)
		}

		reloaded, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to reload config: %v", err)
		}

		// Find the partner admin grant
		var found *config.SyncGrant
		for i, g := range reloaded.RTMX.Sync.Grants {
			if g.Grantee == "partner" && g.Role == "admin" {
				found = &reloaded.RTMX.Sync.Grants[i]
				break
			}
		}
		if found == nil {
			t.Fatal("Partner admin grant not found")
		}
		if len(found.Constraints.Categories) != 2 {
			t.Fatalf("Expected 2 categories, got %d", len(found.Constraints.Categories))
		}
		if found.Constraints.Categories[0] != "AUTH" || found.Constraints.Categories[1] != "API" {
			t.Errorf("Expected categories [AUTH, API], got %v", found.Constraints.Categories)
		}
	})

	t.Run("requirement_ids_constraint_stores_correctly", func(t *testing.T) {
		grantRole = "requirement_editor"
		grantCategories = nil
		grantIDs = []string{"REQ-001", "REQ-002", "REQ-003"}
		grantExcludeCategories = nil
		grantExpiresAt = ""

		buf := new(bytes.Buffer)
		cmd := grantCreateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		if err := cmd.RunE(cmd, []string{"upstream"}); err != nil {
			t.Fatalf("grant create with --ids failed: %v", err)
		}

		reloaded, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to reload config: %v", err)
		}

		// Find the upstream requirement_editor grant
		var found *config.SyncGrant
		for i, g := range reloaded.RTMX.Sync.Grants {
			if g.Grantee == "upstream" && g.Role == "requirement_editor" {
				found = &reloaded.RTMX.Sync.Grants[i]
				break
			}
		}
		if found == nil {
			t.Fatal("Upstream requirement_editor grant not found")
		}
		if len(found.Constraints.RequirementIDs) != 3 {
			t.Fatalf("Expected 3 requirement IDs, got %d", len(found.Constraints.RequirementIDs))
		}
		if found.Constraints.RequirementIDs[0] != "REQ-001" {
			t.Errorf("Expected first ID 'REQ-001', got %q", found.Constraints.RequirementIDs[0])
		}
	})

	t.Run("exclude_categories_constraint_stores_correctly", func(t *testing.T) {
		grantRole = "status_observer"
		grantCategories = nil
		grantIDs = nil
		grantExcludeCategories = []string{"SECRET", "INTERNAL"}
		grantExpiresAt = ""

		buf := new(bytes.Buffer)
		cmd := grantCreateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		if err := cmd.RunE(cmd, []string{"partner"}); err != nil {
			t.Fatalf("grant create with --exclude failed: %v", err)
		}

		reloaded, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to reload config: %v", err)
		}

		// Find the partner status_observer grant with exclude
		var found *config.SyncGrant
		for i, g := range reloaded.RTMX.Sync.Grants {
			if g.Grantee == "partner" && g.Role == "status_observer" && len(g.Constraints.ExcludeCategories) > 0 {
				found = &reloaded.RTMX.Sync.Grants[i]
				break
			}
		}
		if found == nil {
			t.Fatal("Partner status_observer grant with exclude not found")
		}
		if len(found.Constraints.ExcludeCategories) != 2 {
			t.Fatalf("Expected 2 exclude categories, got %d", len(found.Constraints.ExcludeCategories))
		}
		if found.Constraints.ExcludeCategories[0] != "SECRET" {
			t.Errorf("Expected first exclude 'SECRET', got %q", found.Constraints.ExcludeCategories[0])
		}
	})

	t.Run("duplicate_active_grant_returns_error", func(t *testing.T) {
		// upstream already has a status_observer grant from the first test
		grantRole = "status_observer"
		grantCategories = nil
		grantIDs = nil
		grantExcludeCategories = nil
		grantExpiresAt = ""

		err := grantCreateCmd.RunE(grantCreateCmd, []string{"upstream"})
		if err == nil {
			t.Fatal("Expected error for duplicate active grant")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("Expected 'already exists' error, got: %v", err)
		}
	})

	t.Run("all_four_roles_accepted", func(t *testing.T) {
		roles := []string{"dependency_viewer", "status_observer", "requirement_editor", "admin"}
		visibilities := []string{"hash_only", "shadow", "full", "full"}

		// Use a fresh config for this sub-test
		tmpDir2, err := os.MkdirTemp("", "rtmx-grant-roles-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir2) }()

		rtmxDir2 := filepath.Join(tmpDir2, ".rtmx")
		if err := os.MkdirAll(rtmxDir2, 0755); err != nil {
			t.Fatalf("Failed to create .rtmx dir: %v", err)
		}

		cfg := config.DefaultConfig()
		if cfg.RTMX.Sync.Remotes == nil {
			cfg.RTMX.Sync.Remotes = make(map[string]config.SyncRemote)
		}
		cfg.RTMX.Sync.Remotes["target"] = config.SyncRemote{
			Repo:     "org/repo",
			Database: ".rtmx/database.csv",
		}
		configPath2 := filepath.Join(rtmxDir2, "config.yaml")
		if err := cfg.Save(configPath2); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		origDir, _ := os.Getwd()
		if err := os.Chdir(tmpDir2); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		defer func() { _ = os.Chdir(origDir) }()

		for i, role := range roles {
			buf := new(bytes.Buffer)
			cmd := grantCreateCmd
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			grantRole = role
			grantCategories = nil
			grantIDs = nil
			grantExcludeCategories = nil
			grantExpiresAt = ""

			if err := cmd.RunE(cmd, []string{"target"}); err != nil {
				t.Fatalf("grant create with role %q failed: %v", role, err)
			}

			out := buf.String()
			if !strings.Contains(out, "visibility: "+visibilities[i]) {
				t.Errorf("Role %q: expected visibility %q in output, got: %s", role, visibilities[i], out)
			}
		}
	})
}

func TestGrantList(t *testing.T) {
	rtmx.Req(t, "REQ-GO-079")

	t.Run("empty_grants_shows_info_message", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "rtmx-grant-list-empty")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		if err := os.MkdirAll(rtmxDir, 0755); err != nil {
			t.Fatalf("Failed to create .rtmx dir: %v", err)
		}

		cfg := config.DefaultConfig()
		configPath := filepath.Join(rtmxDir, "config.yaml")
		if err := cfg.Save(configPath); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		origDir, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		defer func() { _ = os.Chdir(origDir) }()

		buf := new(bytes.Buffer)
		cmd := grantListCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		if err := cmd.RunE(cmd, nil); err != nil {
			t.Fatalf("grant list failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "No grants configured") {
			t.Errorf("Expected 'No grants configured' message, got: %s", out)
		}
		if !strings.Contains(out, "rtmx grant create") {
			t.Errorf("Expected usage hint in output, got: %s", out)
		}
	})

	t.Run("lists_existing_grants_with_table", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "rtmx-grant-list-table")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		if err := os.MkdirAll(rtmxDir, 0755); err != nil {
			t.Fatalf("Failed to create .rtmx dir: %v", err)
		}

		cfg := config.DefaultConfig()
		cfg.RTMX.Sync.Grants = []config.SyncGrant{
			{
				ID:        "grant-upstream-1000",
				Grantee:   "upstream",
				Role:      "status_observer",
				CreatedAt: "2026-01-01",
				Constraints: config.GrantConstraint{
					Categories: []string{"AUTH", "API"},
				},
			},
			{
				ID:        "grant-partner-2000",
				Grantee:   "partner",
				Role:      "admin",
				CreatedAt: "2026-01-15",
				Constraints: config.GrantConstraint{
					ExpiresAt: "2027-06-30",
				},
			},
			{
				ID:        "grant-expired-3000",
				Grantee:   "old-partner",
				Role:      "dependency_viewer",
				CreatedAt: "2024-01-01",
				Constraints: config.GrantConstraint{
					ExpiresAt: "2024-06-01",
				},
			},
		}

		configPath := filepath.Join(rtmxDir, "config.yaml")
		if err := cfg.Save(configPath); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		origDir, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		defer func() { _ = os.Chdir(origDir) }()

		buf := new(bytes.Buffer)
		cmd := grantListCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		if err := cmd.RunE(cmd, nil); err != nil {
			t.Fatalf("grant list failed: %v", err)
		}

		out := buf.String()

		// Check table headers
		if !strings.Contains(out, "ID") {
			t.Errorf("Expected 'ID' header, got: %s", out)
		}
		if !strings.Contains(out, "GRANTEE") {
			t.Errorf("Expected 'GRANTEE' header, got: %s", out)
		}
		if !strings.Contains(out, "ROLE") {
			t.Errorf("Expected 'ROLE' header, got: %s", out)
		}
		if !strings.Contains(out, "STATUS") {
			t.Errorf("Expected 'STATUS' header, got: %s", out)
		}

		// Check grant entries
		if !strings.Contains(out, "grant-upstream-1000") {
			t.Errorf("Expected upstream grant ID in list, got: %s", out)
		}
		if !strings.Contains(out, "grant-partner-2000") {
			t.Errorf("Expected partner grant ID in list, got: %s", out)
		}
		if !strings.Contains(out, "upstream") {
			t.Errorf("Expected 'upstream' grantee in list, got: %s", out)
		}
		if !strings.Contains(out, "status_observer") {
			t.Errorf("Expected 'status_observer' role in list, got: %s", out)
		}

		// Check constraints display
		if !strings.Contains(out, "categories=") {
			t.Errorf("Expected categories constraint display, got: %s", out)
		}
		if !strings.Contains(out, "expires=2027-06-30") {
			t.Errorf("Expected expires constraint display, got: %s", out)
		}

		// Check expired grant shows "expired" status
		if !strings.Contains(out, "expired") {
			t.Errorf("Expected 'expired' status for old grant, got: %s", out)
		}
		// Check active grants show "active" status
		if !strings.Contains(out, "active") {
			t.Errorf("Expected 'active' status for current grants, got: %s", out)
		}
	})

	t.Run("grant_with_no_constraints_shows_dash", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "rtmx-grant-list-dash")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		if err := os.MkdirAll(rtmxDir, 0755); err != nil {
			t.Fatalf("Failed to create .rtmx dir: %v", err)
		}

		cfg := config.DefaultConfig()
		cfg.RTMX.Sync.Grants = []config.SyncGrant{
			{
				ID:        "grant-noconstrain-1000",
				Grantee:   "simple",
				Role:      "admin",
				CreatedAt: "2026-01-01",
			},
		}

		configPath := filepath.Join(rtmxDir, "config.yaml")
		if err := cfg.Save(configPath); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		origDir, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		defer func() { _ = os.Chdir(origDir) }()

		buf := new(bytes.Buffer)
		cmd := grantListCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		if err := cmd.RunE(cmd, nil); err != nil {
			t.Fatalf("grant list failed: %v", err)
		}

		out := buf.String()
		// The constraints column should show "-" for no constraints
		lines := strings.Split(out, "\n")
		foundDash := false
		for _, line := range lines {
			if strings.Contains(line, "grant-noconstrain-1000") && strings.HasSuffix(strings.TrimSpace(line), "-") {
				foundDash = true
				break
			}
		}
		if !foundDash {
			t.Errorf("Expected '-' for no constraints in grant line, got: %s", out)
		}
	})
}

func TestGrantRevoke(t *testing.T) {
	rtmx.Req(t, "REQ-GO-079")

	t.Run("revoke_removes_grant_from_config", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "rtmx-grant-revoke")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		if err := os.MkdirAll(rtmxDir, 0755); err != nil {
			t.Fatalf("Failed to create .rtmx dir: %v", err)
		}

		cfg := config.DefaultConfig()
		cfg.RTMX.Sync.Grants = []config.SyncGrant{
			{
				ID:        "grant-upstream-1000",
				Grantee:   "upstream",
				Role:      "status_observer",
				CreatedAt: "2026-01-01",
			},
			{
				ID:        "grant-partner-2000",
				Grantee:   "partner",
				Role:      "admin",
				CreatedAt: "2026-01-15",
			},
		}

		configPath := filepath.Join(rtmxDir, "config.yaml")
		if err := cfg.Save(configPath); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		origDir, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		defer func() { _ = os.Chdir(origDir) }()

		buf := new(bytes.Buffer)
		cmd := grantRevokeCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		if err := cmd.RunE(cmd, []string{"grant-upstream-1000"}); err != nil {
			t.Fatalf("grant revoke failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Revoked grant grant-upstream-1000") {
			t.Errorf("Expected revoke confirmation, got: %s", out)
		}

		// Verify grant was removed from config
		reloaded, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to reload config: %v", err)
		}
		if len(reloaded.RTMX.Sync.Grants) != 1 {
			t.Fatalf("Expected 1 remaining grant, got %d", len(reloaded.RTMX.Sync.Grants))
		}
		if reloaded.RTMX.Sync.Grants[0].ID != "grant-partner-2000" {
			t.Errorf("Expected remaining grant to be 'grant-partner-2000', got %q", reloaded.RTMX.Sync.Grants[0].ID)
		}
	})

	t.Run("revoke_nonexistent_returns_error", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "rtmx-grant-revoke-noexist")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		if err := os.MkdirAll(rtmxDir, 0755); err != nil {
			t.Fatalf("Failed to create .rtmx dir: %v", err)
		}

		cfg := config.DefaultConfig()
		cfg.RTMX.Sync.Grants = []config.SyncGrant{
			{
				ID:        "grant-upstream-1000",
				Grantee:   "upstream",
				Role:      "status_observer",
				CreatedAt: "2026-01-01",
			},
		}

		configPath := filepath.Join(rtmxDir, "config.yaml")
		if err := cfg.Save(configPath); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		origDir, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		defer func() { _ = os.Chdir(origDir) }()

		err = grantRevokeCmd.RunE(grantRevokeCmd, []string{"grant-nonexistent-9999"})
		if err == nil {
			t.Fatal("Expected error for nonexistent grant ID")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}

		// Verify config was not modified
		reloaded, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to reload config: %v", err)
		}
		if len(reloaded.RTMX.Sync.Grants) != 1 {
			t.Errorf("Expected grants unchanged after failed revoke, got %d", len(reloaded.RTMX.Sync.Grants))
		}
	})

	t.Run("revoke_last_grant_leaves_empty_list", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "rtmx-grant-revoke-last")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		if err := os.MkdirAll(rtmxDir, 0755); err != nil {
			t.Fatalf("Failed to create .rtmx dir: %v", err)
		}

		cfg := config.DefaultConfig()
		cfg.RTMX.Sync.Grants = []config.SyncGrant{
			{
				ID:        "grant-only-1000",
				Grantee:   "sole",
				Role:      "admin",
				CreatedAt: "2026-01-01",
			},
		}

		configPath := filepath.Join(rtmxDir, "config.yaml")
		if err := cfg.Save(configPath); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		origDir, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		defer func() { _ = os.Chdir(origDir) }()

		buf := new(bytes.Buffer)
		cmd := grantRevokeCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		if err := cmd.RunE(cmd, []string{"grant-only-1000"}); err != nil {
			t.Fatalf("grant revoke failed: %v", err)
		}

		reloaded, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to reload config: %v", err)
		}
		if len(reloaded.RTMX.Sync.Grants) != 0 {
			t.Errorf("Expected 0 grants after revoking last one, got %d", len(reloaded.RTMX.Sync.Grants))
		}
	})
}

package sync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestGrantDelegation(t *testing.T) {
	rtmx.Req(t, "REQ-GO-036")

	t.Run("VisibilityForRole", func(t *testing.T) {
		tests := []struct {
			role     string
			expected string
		}{
			{RoleAdmin, "full"},
			{RoleRequirementEditor, "full"},
			{RoleStatusObserver, "shadow"},
			{RoleDependencyViewer, "hash_only"},
			{"unknown", "hash_only"},
		}
		for _, tt := range tests {
			if got := VisibilityForRole(tt.role); got != tt.expected {
				t.Errorf("VisibilityForRole(%q) = %q, want %q", tt.role, got, tt.expected)
			}
		}
	})

	t.Run("IsGrantActive_no_expiry", func(t *testing.T) {
		grant := config.SyncGrant{ID: "g1", Grantee: "x", Role: RoleAdmin}
		if !IsGrantActive(grant) {
			t.Error("grant with no expiry should be active")
		}
	})

	t.Run("IsGrantActive_future_expiry", func(t *testing.T) {
		grant := config.SyncGrant{
			ID: "g1", Grantee: "x", Role: RoleAdmin,
			Constraints: config.GrantConstraint{ExpiresAt: "2099-12-31"},
		}
		if !IsGrantActive(grant) {
			t.Error("grant with future expiry should be active")
		}
	})

	t.Run("IsGrantActive_past_expiry", func(t *testing.T) {
		grant := config.SyncGrant{
			ID: "g1", Grantee: "x", Role: RoleAdmin,
			Constraints: config.GrantConstraint{ExpiresAt: "2020-01-01"},
		}
		if IsGrantActive(grant) {
			t.Error("grant with past expiry should be inactive")
		}
	})

	t.Run("IsGrantActive_malformed_date", func(t *testing.T) {
		grant := config.SyncGrant{
			ID: "g1", Grantee: "x", Role: RoleAdmin,
			Constraints: config.GrantConstraint{ExpiresAt: "not-a-date"},
		}
		if IsGrantActive(grant) {
			t.Error("grant with malformed date should be inactive")
		}
	})

	t.Run("ConstraintAllows_no_constraints", func(t *testing.T) {
		constraint := config.GrantConstraint{}
		req := &database.Requirement{ReqID: "REQ-001", Category: "AUTH"}
		if !ConstraintAllows(constraint, req) {
			t.Error("empty constraints should allow all requirements")
		}
	})

	t.Run("ConstraintAllows_category_whitelist_match", func(t *testing.T) {
		constraint := config.GrantConstraint{Categories: []string{"AUTH", "API"}}
		req := &database.Requirement{ReqID: "REQ-001", Category: "AUTH"}
		if !ConstraintAllows(constraint, req) {
			t.Error("matching category should be allowed")
		}
	})

	t.Run("ConstraintAllows_category_whitelist_no_match", func(t *testing.T) {
		constraint := config.GrantConstraint{Categories: []string{"AUTH", "API"}}
		req := &database.Requirement{ReqID: "REQ-001", Category: "CLI"}
		if ConstraintAllows(constraint, req) {
			t.Error("non-matching category should be rejected")
		}
	})

	t.Run("ConstraintAllows_id_whitelist_match", func(t *testing.T) {
		constraint := config.GrantConstraint{RequirementIDs: []string{"REQ-001", "REQ-002"}}
		req := &database.Requirement{ReqID: "REQ-001", Category: "AUTH"}
		if !ConstraintAllows(constraint, req) {
			t.Error("matching ID should be allowed")
		}
	})

	t.Run("ConstraintAllows_id_whitelist_no_match", func(t *testing.T) {
		constraint := config.GrantConstraint{RequirementIDs: []string{"REQ-001", "REQ-002"}}
		req := &database.Requirement{ReqID: "REQ-999", Category: "AUTH"}
		if ConstraintAllows(constraint, req) {
			t.Error("non-matching ID should be rejected")
		}
	})

	t.Run("ConstraintAllows_exclude_category", func(t *testing.T) {
		constraint := config.GrantConstraint{ExcludeCategories: []string{"SECRET"}}
		req := &database.Requirement{ReqID: "REQ-001", Category: "SECRET"}
		if ConstraintAllows(constraint, req) {
			t.Error("excluded category should be rejected")
		}
	})

	t.Run("ConstraintAllows_exclude_does_not_affect_other", func(t *testing.T) {
		constraint := config.GrantConstraint{ExcludeCategories: []string{"SECRET"}}
		req := &database.Requirement{ReqID: "REQ-001", Category: "AUTH"}
		if !ConstraintAllows(constraint, req) {
			t.Error("non-excluded category should be allowed")
		}
	})

	t.Run("FindGrant_exists", func(t *testing.T) {
		grants := []config.SyncGrant{
			{ID: "g1", Grantee: "a"},
			{ID: "g2", Grantee: "b"},
		}
		g := FindGrant(grants, "g2")
		if g == nil || g.Grantee != "b" {
			t.Error("should find grant g2")
		}
	})

	t.Run("FindGrant_not_found", func(t *testing.T) {
		grants := []config.SyncGrant{{ID: "g1"}}
		if FindGrant(grants, "g99") != nil {
			t.Error("should return nil for missing grant")
		}
	})

	t.Run("FindGrantByGrantee", func(t *testing.T) {
		grants := []config.SyncGrant{
			{ID: "g1", Grantee: "upstream", Role: RoleAdmin},
			{ID: "g2", Grantee: "other", Role: RoleAdmin},
			{ID: "g3", Grantee: "upstream", Role: RoleStatusObserver},
		}
		found := FindGrantByGrantee(grants, "upstream")
		if len(found) != 2 {
			t.Errorf("expected 2 grants for upstream, got %d", len(found))
		}
	})

	t.Run("ValidateNewGrant_valid", func(t *testing.T) {
		err := ValidateNewGrant(nil, "upstream", RoleAdmin)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("ValidateNewGrant_invalid_role", func(t *testing.T) {
		err := ValidateNewGrant(nil, "upstream", "superuser")
		if err == nil {
			t.Fatal("expected error for invalid role")
		}
	})

	t.Run("ValidateNewGrant_duplicate", func(t *testing.T) {
		existing := []config.SyncGrant{
			{ID: "g1", Grantee: "upstream", Role: RoleAdmin},
		}
		err := ValidateNewGrant(existing, "upstream", RoleAdmin)
		if err == nil {
			t.Fatal("expected error for duplicate grant")
		}
	})

	t.Run("ValidateNewGrant_expired_duplicate_ok", func(t *testing.T) {
		existing := []config.SyncGrant{
			{ID: "g1", Grantee: "upstream", Role: RoleAdmin,
				Constraints: config.GrantConstraint{ExpiresAt: "2020-01-01"}},
		}
		err := ValidateNewGrant(existing, "upstream", RoleAdmin)
		if err != nil {
			t.Errorf("expired duplicate should be allowed: %v", err)
		}
	})

	t.Run("GenerateGrantID_format", func(t *testing.T) {
		id := GenerateGrantID("upstream")
		if len(id) == 0 {
			t.Error("grant ID should not be empty")
		}
		if id[:14] != "grant-upstream" {
			t.Errorf("expected grant-upstream prefix, got %q", id)
		}
	})

	t.Run("grant_persists_to_config", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "rtmx-grant-test")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		if err := os.MkdirAll(rtmxDir, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}

		cfg := config.DefaultConfig()
		grant := config.SyncGrant{
			ID:        "grant-test-001",
			Grantee:   "upstream",
			Role:      RoleStatusObserver,
			CreatedAt: "2026-03-25",
			CreatedBy: "test@rtmx.ai",
			Constraints: config.GrantConstraint{
				Categories: []string{"AUTH", "API"},
			},
		}
		cfg.RTMX.Sync.Grants = append(cfg.RTMX.Sync.Grants, grant)

		configPath := filepath.Join(rtmxDir, "config.yaml")
		if err := cfg.Save(configPath); err != nil {
			t.Fatalf("failed to save config: %v", err)
		}

		reloaded, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("failed to reload config: %v", err)
		}

		if len(reloaded.RTMX.Sync.Grants) != 1 {
			t.Fatalf("expected 1 grant, got %d", len(reloaded.RTMX.Sync.Grants))
		}
		g := reloaded.RTMX.Sync.Grants[0]
		if g.ID != "grant-test-001" {
			t.Errorf("expected ID grant-test-001, got %q", g.ID)
		}
		if g.Role != RoleStatusObserver {
			t.Errorf("expected role status_observer, got %q", g.Role)
		}
		if len(g.Constraints.Categories) != 2 {
			t.Errorf("expected 2 categories, got %d", len(g.Constraints.Categories))
		}
	})
}

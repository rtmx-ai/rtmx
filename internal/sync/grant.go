package sync

import (
	"fmt"
	"time"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
)

// Role constants for grant delegation.
const (
	RoleDependencyViewer  = "dependency_viewer"
	RoleStatusObserver    = "status_observer"
	RoleRequirementEditor = "requirement_editor"
	RoleAdmin             = "admin"
)

// ValidRoles is the set of valid grant roles.
var ValidRoles = map[string]bool{
	RoleDependencyViewer:  true,
	RoleStatusObserver:    true,
	RoleRequirementEditor: true,
	RoleAdmin:             true,
}

// VisibilityForRole returns the visibility level for a given role.
func VisibilityForRole(role string) string {
	switch role {
	case RoleAdmin, RoleRequirementEditor:
		return "full"
	case RoleStatusObserver:
		return "shadow"
	case RoleDependencyViewer:
		return "hash_only"
	default:
		return "hash_only"
	}
}

// IsGrantActive returns true if the grant has not expired.
func IsGrantActive(grant config.SyncGrant) bool {
	if grant.Constraints.ExpiresAt == "" {
		return true
	}
	expires, err := time.Parse("2006-01-02", grant.Constraints.ExpiresAt)
	if err != nil {
		return false // Malformed date treated as expired
	}
	return time.Now().Before(expires.Add(24 * time.Hour)) // Inclusive of the expiry day
}

// ConstraintAllows returns true if the grant's constraints permit access to the given requirement.
func ConstraintAllows(constraint config.GrantConstraint, req *database.Requirement) bool {
	// Check category whitelist
	if len(constraint.Categories) > 0 {
		found := false
		for _, cat := range constraint.Categories {
			if cat == req.Category {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check requirement ID whitelist
	if len(constraint.RequirementIDs) > 0 {
		found := false
		for _, id := range constraint.RequirementIDs {
			if id == req.ReqID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check category blacklist
	for _, cat := range constraint.ExcludeCategories {
		if cat == req.Category {
			return false
		}
	}

	return true
}

// FindGrant returns the grant with the given ID, or nil if not found.
func FindGrant(grants []config.SyncGrant, id string) *config.SyncGrant {
	for i := range grants {
		if grants[i].ID == id {
			return &grants[i]
		}
	}
	return nil
}

// FindGrantByGrantee returns all grants for a given grantee alias.
func FindGrantByGrantee(grants []config.SyncGrant, grantee string) []config.SyncGrant {
	var result []config.SyncGrant
	for _, g := range grants {
		if g.Grantee == grantee {
			result = append(result, g)
		}
	}
	return result
}

// ValidateNewGrant checks if a new grant can be created.
func ValidateNewGrant(grants []config.SyncGrant, grantee, role string) error {
	if !ValidRoles[role] {
		return fmt.Errorf("invalid role %q (valid: dependency_viewer, status_observer, requirement_editor, admin)", role)
	}

	// Check for duplicates
	for _, g := range grants {
		if g.Grantee == grantee && g.Role == role && IsGrantActive(g) {
			return fmt.Errorf("active grant already exists for grantee %q with role %q (id: %s)", grantee, role, g.ID)
		}
	}

	return nil
}

// GenerateGrantID creates a unique grant ID based on grantee and timestamp.
func GenerateGrantID(grantee string) string {
	return fmt.Sprintf("grant-%s-%d", grantee, time.Now().UnixMilli())
}

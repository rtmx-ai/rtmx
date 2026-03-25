// Package sync provides cross-repository requirement synchronization.
package sync

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
)

// ShadowRequirement is a local projection of a requirement from another repository.
type ShadowRequirement struct {
	ReqID        string          // Original ID in the remote repo
	RemoteAlias  string          // Config alias (e.g., "rtmx")
	RemoteRepo   string          // Repository identifier (e.g., "rtmx-ai/rtmx")
	Status       database.Status // COMPLETE, PARTIAL, MISSING
	Description  string          // Requirement text
	Phase        int             // Phase number
	Dependencies []string        // Dependencies within the remote repo
	Visibility   string          // "full" (default)
	ResolvedAt   time.Time       // When this shadow was last resolved
}

// Warning represents a non-fatal issue during shadow resolution.
type Warning struct {
	Ref     string // The original dependency reference
	Message string // Human-readable warning
}

func (w Warning) String() string {
	return fmt.Sprintf("%s: %s", w.Ref, w.Message)
}

// ShadowResolver resolves cross-repo dependency references to ShadowRequirements.
type ShadowResolver struct {
	Remotes  map[string]config.SyncRemote
	cache    map[string]*ShadowRequirement // keyed by "alias:req_id"
	dbCache  map[string]*database.Database // keyed by alias
	warnings []Warning
}

// NewShadowResolver creates a resolver with the given remote configuration.
func NewShadowResolver(remotes map[string]config.SyncRemote) *ShadowResolver {
	return &ShadowResolver{
		Remotes: remotes,
		cache:   make(map[string]*ShadowRequirement),
		dbCache: make(map[string]*database.Database),
	}
}

// Warnings returns accumulated warnings from resolution.
func (r *ShadowResolver) Warnings() []Warning {
	return r.warnings
}

// IsResolvable returns true if the dependency string is a cross-repo reference.
func IsResolvable(dep string) bool {
	return strings.HasPrefix(dep, "sync:")
}

// ParseRef splits "sync:ALIAS:REQ-ID" into alias and requirement ID.
func ParseRef(ref string) (alias string, reqID string, err error) {
	if !strings.HasPrefix(ref, "sync:") {
		return "", "", fmt.Errorf("not a cross-repo reference: %q", ref)
	}
	parts := strings.SplitN(ref, ":", 3)
	if len(parts) != 3 || parts[1] == "" || parts[2] == "" {
		return "", "", fmt.Errorf("invalid cross-repo reference format: %q (expected sync:ALIAS:REQ-ID)", ref)
	}
	return parts[1], parts[2], nil
}

// Resolve parses a "sync:ALIAS:REQ-ID" reference and returns the shadow requirement.
func (r *ShadowResolver) Resolve(ref string) (*ShadowRequirement, error) {
	alias, reqID, err := ParseRef(ref)
	if err != nil {
		return nil, err
	}

	cacheKey := alias + ":" + reqID
	if cached, ok := r.cache[cacheKey]; ok {
		return cached, nil
	}

	remote, ok := r.Remotes[alias]
	if !ok {
		return nil, fmt.Errorf("unknown remote alias %q", alias)
	}

	if remote.Path == "" {
		return nil, fmt.Errorf("remote %q has no local path configured (use rtmx remote add %s --repo %s --path PATH)", alias, alias, remote.Repo)
	}

	db, err := r.loadRemoteDB(alias, remote)
	if err != nil {
		return nil, fmt.Errorf("failed to load remote database for %q: %w", alias, err)
	}

	req := db.Get(reqID)
	if req == nil {
		return nil, fmt.Errorf("requirement %q not found in remote %q", reqID, alias)
	}

	shadow := &ShadowRequirement{
		ReqID:        req.ReqID,
		RemoteAlias:  alias,
		RemoteRepo:   remote.Repo,
		Status:       req.Status,
		Description:  req.RequirementText,
		Phase:        req.Phase,
		Dependencies: req.Dependencies.Slice(),
		Visibility:   "full",
		ResolvedAt:   time.Now(),
	}

	r.cache[cacheKey] = shadow
	return shadow, nil
}

// ResolveAll resolves all cross-repo dependencies in a database.
// Returns resolved shadows and accumulated warnings.
func (r *ShadowResolver) ResolveAll(db *database.Database) []*ShadowRequirement {
	r.warnings = nil
	seen := make(map[string]bool)
	var shadows []*ShadowRequirement

	for _, req := range db.All() {
		for dep := range req.Dependencies {
			if !IsResolvable(dep) || seen[dep] {
				continue
			}
			seen[dep] = true

			shadow, err := r.Resolve(dep)
			if err != nil {
				r.warnings = append(r.warnings, Warning{
					Ref:     dep,
					Message: err.Error(),
				})
				continue
			}
			shadows = append(shadows, shadow)
		}
	}

	return shadows
}

// IsShadowBlocking returns true if the given cross-repo ref is blocking (incomplete).
// Returns false if the shadow cannot be resolved (permissive).
func (r *ShadowResolver) IsShadowBlocking(ref string) bool {
	shadow, err := r.Resolve(ref)
	if err != nil {
		// Unresolvable shadows do not block
		return false
	}
	return shadow.Status.IsIncomplete()
}

func (r *ShadowResolver) loadRemoteDB(alias string, remote config.SyncRemote) (*database.Database, error) {
	if cached, ok := r.dbCache[alias]; ok {
		return cached, nil
	}

	dbPath := filepath.Join(remote.Path, remote.Database)
	db, err := database.Load(dbPath)
	if err != nil {
		return nil, err
	}

	r.dbCache[alias] = db
	return db, nil
}

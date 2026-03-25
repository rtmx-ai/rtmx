package sync

import (
	"strings"
	"testing"
	"time"

	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
)

func TestCRDTMerge(t *testing.T) {
	rtmx.Req(t, "REQ-GO-042")

	t.Run("DatabaseToUpdates_serializes_all_fields", func(t *testing.T) {
		db := database.NewDatabase()
		req := database.NewRequirement("REQ-001")
		req.Category = "AUTH"
		req.Subcategory = "Login"
		req.RequirementText = "User can log in"
		req.Status = database.StatusComplete
		_ = db.Add(req)

		updates := DatabaseToUpdates(db)
		if len(updates) != 1 {
			t.Fatalf("expected 1 update, got %d", len(updates))
		}
		u := updates[0]
		if u.ReqID != "REQ-001" {
			t.Errorf("expected REQ-001, got %q", u.ReqID)
		}
		if u.Fields["category"] != "AUTH" {
			t.Errorf("expected category AUTH, got %q", u.Fields["category"])
		}
		if u.Fields["status"] != "COMPLETE" {
			t.Errorf("expected status COMPLETE, got %q", u.Fields["status"])
		}
		if u.Action != "updated" {
			t.Errorf("expected action 'updated', got %q", u.Action)
		}
	})

	t.Run("DatabaseToUpdates_empty_database", func(t *testing.T) {
		db := database.NewDatabase()
		updates := DatabaseToUpdates(db)
		if len(updates) != 0 {
			t.Errorf("expected 0 updates for empty db, got %d", len(updates))
		}
	})

	t.Run("ApplyUpdates_add_new_requirement", func(t *testing.T) {
		db := database.NewDatabase()
		updates := []RequirementUpdate{
			{
				ReqID:  "REQ-NEW-001",
				Action: "added",
				Fields: map[string]string{
					"category":         "API",
					"requirement_text": "New endpoint",
					"status":           "MISSING",
				},
			},
		}

		result := ApplyUpdates(db, updates)
		if len(result.Added) != 1 || result.Added[0] != "REQ-NEW-001" {
			t.Errorf("expected REQ-NEW-001 added, got %v", result.Added)
		}

		req := db.Get("REQ-NEW-001")
		if req == nil {
			t.Fatal("requirement not found after add")
		}
		if req.Category != "API" {
			t.Errorf("expected category API, got %q", req.Category)
		}
		if req.RequirementText != "New endpoint" {
			t.Errorf("expected text 'New endpoint', got %q", req.RequirementText)
		}
	})

	t.Run("ApplyUpdates_update_existing", func(t *testing.T) {
		db := database.NewDatabase()
		req := database.NewRequirement("REQ-001")
		req.Status = database.StatusMissing
		_ = db.Add(req)

		updates := []RequirementUpdate{
			{
				ReqID:  "REQ-001",
				Action: "updated",
				Fields: map[string]string{"status": "COMPLETE"},
			},
		}

		result := ApplyUpdates(db, updates)
		if len(result.Updated) != 1 {
			t.Errorf("expected 1 updated, got %d", len(result.Updated))
		}

		updated := db.Get("REQ-001")
		if !updated.Status.IsComplete() {
			t.Errorf("expected COMPLETE, got %v", updated.Status)
		}
	})

	t.Run("ApplyUpdates_update_nonexistent_adds_it", func(t *testing.T) {
		db := database.NewDatabase()
		updates := []RequirementUpdate{
			{
				ReqID:  "REQ-REMOTE-001",
				Action: "updated",
				Fields: map[string]string{"status": "COMPLETE", "category": "SYNC"},
			},
		}

		result := ApplyUpdates(db, updates)
		if len(result.Added) != 1 {
			t.Errorf("expected 1 added (for nonexistent update), got %d", len(result.Added))
		}
		if db.Get("REQ-REMOTE-001") == nil {
			t.Error("requirement should have been added")
		}
	})

	t.Run("ApplyUpdates_remove_existing", func(t *testing.T) {
		db := database.NewDatabase()
		_ = db.Add(database.NewRequirement("REQ-001"))

		updates := []RequirementUpdate{
			{ReqID: "REQ-001", Action: "removed"},
		}

		result := ApplyUpdates(db, updates)
		if len(result.Removed) != 1 {
			t.Errorf("expected 1 removed, got %d", len(result.Removed))
		}
		if db.Get("REQ-001") != nil {
			t.Error("requirement should have been removed")
		}
	})

	t.Run("ApplyUpdates_remove_nonexistent_is_noop", func(t *testing.T) {
		db := database.NewDatabase()
		updates := []RequirementUpdate{
			{ReqID: "REQ-GHOST", Action: "removed"},
		}

		result := ApplyUpdates(db, updates)
		if len(result.Removed) != 0 {
			t.Errorf("expected 0 removed for nonexistent, got %d", len(result.Removed))
		}
	})

	t.Run("ApplyUpdates_add_duplicate_treated_as_update", func(t *testing.T) {
		db := database.NewDatabase()
		req := database.NewRequirement("REQ-001")
		req.Category = "OLD"
		_ = db.Add(req)

		updates := []RequirementUpdate{
			{
				ReqID:  "REQ-001",
				Action: "added",
				Fields: map[string]string{"category": "NEW"},
			},
		}

		result := ApplyUpdates(db, updates)
		if len(result.Updated) != 1 {
			t.Errorf("duplicate add should be treated as update, got added=%d updated=%d", len(result.Added), len(result.Updated))
		}
		if db.Get("REQ-001").Category != "NEW" {
			t.Error("category should have been updated")
		}
	})

	t.Run("ApplyUpdates_dependencies_parsed", func(t *testing.T) {
		db := database.NewDatabase()
		updates := []RequirementUpdate{
			{
				ReqID:  "REQ-001",
				Action: "added",
				Fields: map[string]string{
					"dependencies": "REQ-002|REQ-003",
					"blocks":       "REQ-004",
				},
			},
		}

		ApplyUpdates(db, updates)
		req := db.Get("REQ-001")
		if !req.Dependencies.Contains("REQ-002") || !req.Dependencies.Contains("REQ-003") {
			t.Errorf("dependencies not parsed correctly: %v", req.Dependencies)
		}
		if !req.Blocks.Contains("REQ-004") {
			t.Errorf("blocks not parsed correctly: %v", req.Blocks)
		}
	})

	t.Run("SyncResult_Summary", func(t *testing.T) {
		tests := []struct {
			name     string
			result   SyncResult
			contains string
		}{
			{"empty", SyncResult{}, "No changes"},
			{"added", SyncResult{Added: []string{"REQ-1"}}, "1 added"},
			{"mixed", SyncResult{Added: []string{"REQ-1"}, Updated: []string{"REQ-2", "REQ-3"}}, "2 updated"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				s := tt.result.Summary()
				if !strings.Contains(s, tt.contains) {
					t.Errorf("expected %q in summary, got %q", tt.contains, s)
				}
			})
		}
	})

	t.Run("SyncResult_HasChanges", func(t *testing.T) {
		if (&SyncResult{}).HasChanges() {
			t.Error("empty result should have no changes")
		}
		if !(&SyncResult{Added: []string{"x"}}).HasChanges() {
			t.Error("result with added should have changes")
		}
	})

	t.Run("EncodeDecode_roundtrip", func(t *testing.T) {
		msg := &SyncMessage{
			Type: MessageTypePush,
			Room: "rtmx-ai/rtmx-go",
			Updates: []RequirementUpdate{
				{
					ReqID:     "REQ-001",
					Action:    "updated",
					Fields:    map[string]string{"status": "COMPLETE"},
					Timestamp: time.Now(),
				},
			},
			Timestamp: time.Now(),
		}

		data, err := EncodeSyncMessage(msg)
		if err != nil {
			t.Fatalf("encode failed: %v", err)
		}

		decoded, err := DecodeSyncMessage(data)
		if err != nil {
			t.Fatalf("decode failed: %v", err)
		}

		if decoded.Type != MessageTypePush {
			t.Errorf("expected type push, got %q", decoded.Type)
		}
		if decoded.Room != "rtmx-ai/rtmx-go" {
			t.Errorf("expected room rtmx-ai/rtmx-go, got %q", decoded.Room)
		}
		if len(decoded.Updates) != 1 {
			t.Errorf("expected 1 update, got %d", len(decoded.Updates))
		}
	})

	t.Run("DecodeSyncMessage_invalid_json", func(t *testing.T) {
		_, err := DecodeSyncMessage([]byte("not json"))
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

package sync

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rtmx-ai/rtmx/internal/database"
)

// MessageType identifies the type of sync protocol message.
type MessageType string

const (
	MessageTypePush   MessageType = "push"
	MessageTypePull   MessageType = "pull"
	MessageTypeUpdate MessageType = "update"
	MessageTypeAck    MessageType = "ack"
	MessageTypeError  MessageType = "error"
)

// SyncMessage is the wire format for sync protocol messages.
type SyncMessage struct {
	Type      MessageType          `json:"type"`
	Room      string               `json:"room,omitempty"`
	Updates   []RequirementUpdate  `json:"updates,omitempty"`
	Error     string               `json:"error,omitempty"`
	Timestamp time.Time            `json:"timestamp"`
}

// RequirementUpdate represents a single requirement change.
type RequirementUpdate struct {
	ReqID     string            `json:"req_id"`
	Action    string            `json:"action"` // "added", "updated", "removed"
	Fields    map[string]string `json:"fields,omitempty"`
	Source    string            `json:"source,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// SyncResult summarizes the outcome of a push or pull operation.
type SyncResult struct {
	Added    []string
	Updated  []string
	Removed  []string
	Errors   []string
}

// HasChanges returns true if any requirements were affected.
func (r *SyncResult) HasChanges() bool {
	return len(r.Added) > 0 || len(r.Updated) > 0 || len(r.Removed) > 0
}

// Summary returns a human-readable summary of changes.
func (r *SyncResult) Summary() string {
	if !r.HasChanges() {
		return "No changes"
	}
	parts := make([]string, 0, 3)
	if len(r.Added) > 0 {
		parts = append(parts, fmt.Sprintf("%d added", len(r.Added)))
	}
	if len(r.Updated) > 0 {
		parts = append(parts, fmt.Sprintf("%d updated", len(r.Updated)))
	}
	if len(r.Removed) > 0 {
		parts = append(parts, fmt.Sprintf("%d removed", len(r.Removed)))
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}

// DatabaseToUpdates converts a database to a list of push updates.
func DatabaseToUpdates(db *database.Database) []RequirementUpdate {
	now := time.Now()
	var updates []RequirementUpdate
	for _, req := range db.All() {
		fields := map[string]string{
			"category":          req.Category,
			"subcategory":       req.Subcategory,
			"requirement_text":  req.RequirementText,
			"target_value":      req.TargetValue,
			"status":            string(req.Status),
			"priority":          string(req.Priority),
			"phase":             fmt.Sprintf("%d", req.Phase),
			"test_module":       req.TestModule,
			"test_function":     req.TestFunction,
			"validation_method": req.ValidationMethod,
			"dependencies":      req.Dependencies.String(),
			"blocks":            req.Blocks.String(),
		}
		if req.Notes != "" {
			fields["notes"] = req.Notes
		}
		if req.EffortWeeks > 0 {
			fields["effort_weeks"] = fmt.Sprintf("%.1f", req.EffortWeeks)
		}

		updates = append(updates, RequirementUpdate{
			ReqID:     req.ReqID,
			Action:    "updated",
			Fields:    fields,
			Timestamp: now,
		})
	}
	return updates
}

// ApplyUpdates applies a list of updates to a database, returning a SyncResult.
func ApplyUpdates(db *database.Database, updates []RequirementUpdate) *SyncResult {
	result := &SyncResult{}

	for _, u := range updates {
		switch u.Action {
		case "added":
			if existing := db.Get(u.ReqID); existing != nil {
				applyFields(existing, u.Fields)
				result.Updated = append(result.Updated, u.ReqID)
			} else {
				req := database.NewRequirement(u.ReqID)
				applyFields(req, u.Fields)
				if err := db.Add(req); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", u.ReqID, err))
				} else {
					result.Added = append(result.Added, u.ReqID)
				}
			}
		case "updated":
			if existing := db.Get(u.ReqID); existing != nil {
				applyFields(existing, u.Fields)
				result.Updated = append(result.Updated, u.ReqID)
			} else {
				req := database.NewRequirement(u.ReqID)
				applyFields(req, u.Fields)
				if err := db.Add(req); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", u.ReqID, err))
				} else {
					result.Added = append(result.Added, u.ReqID)
				}
			}
		case "removed":
			if err := db.Remove(u.ReqID); err == nil {
				result.Removed = append(result.Removed, u.ReqID)
			}
		}
	}

	return result
}

func applyFields(req *database.Requirement, fields map[string]string) {
	if v, ok := fields["category"]; ok {
		req.Category = v
	}
	if v, ok := fields["subcategory"]; ok {
		req.Subcategory = v
	}
	if v, ok := fields["requirement_text"]; ok {
		req.RequirementText = v
	}
	if v, ok := fields["target_value"]; ok {
		req.TargetValue = v
	}
	if v, ok := fields["status"]; ok {
		if s, err := database.ParseStatus(v); err == nil {
			req.Status = s
		}
	}
	if v, ok := fields["priority"]; ok {
		if p, err := database.ParsePriority(v); err == nil {
			req.Priority = p
		}
	}
	if v, ok := fields["dependencies"]; ok {
		req.Dependencies = database.ParseStringSet(v)
	}
	if v, ok := fields["blocks"]; ok {
		req.Blocks = database.ParseStringSet(v)
	}
	if v, ok := fields["notes"]; ok {
		req.Notes = v
	}
	if v, ok := fields["test_module"]; ok {
		req.TestModule = v
	}
	if v, ok := fields["test_function"]; ok {
		req.TestFunction = v
	}
}

// EncodeSyncMessage serializes a SyncMessage to JSON.
func EncodeSyncMessage(msg *SyncMessage) ([]byte, error) {
	return json.Marshal(msg)
}

// DecodeSyncMessage deserializes a SyncMessage from JSON.
func DecodeSyncMessage(data []byte) (*SyncMessage, error) {
	var msg SyncMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to decode sync message: %w", err)
	}
	return &msg, nil
}

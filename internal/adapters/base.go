// Package adapters provides integrations with external services like GitHub and Jira.
package adapters

import (
	"regexp"

	"github.com/rtmx-ai/rtmx/internal/database"
)

// reqIDRegex matches REQ-<CATEGORY>-<NUMBER> where CATEGORY is one or more
// uppercase-alphanumeric segments (e.g., CLI, MCP, E2E, V2, K8S, INFRA-DT,
// MODE-S). The first character of the prefix must be a letter; subsequent
// characters may include digits. The final hyphen-delimited segment is always
// the number.
var reqIDRegex = regexp.MustCompile(`REQ-[A-Z][A-Z0-9]*(-[A-Z0-9]+)*-\d+`)

// ExtractReqID finds the first RTMX requirement ID in text.
// It handles bracketed titles ([REQ-CLI-001]), RTMX: prefixes,
// and bare inline mentions.
func ExtractReqID(text string) string {
	if m := reqIDRegex.FindString(text); m != "" {
		return m
	}
	return ""
}

// ServiceAdapter defines the interface for external service adapters
type ServiceAdapter interface {
	// Name returns the adapter name (e.g., "github", "jira")
	Name() string

	// IsConfigured checks if the adapter is properly configured
	IsConfigured() bool

	// TestConnection tests the connection to the external service
	TestConnection() (success bool, message string)

	// FetchItems fetches items from the external service
	// query can contain service-specific filter parameters
	FetchItems(query map[string]interface{}) ([]ExternalItem, error)

	// GetItem gets a single item by its external ID
	GetItem(externalID string) (*ExternalItem, error)

	// CreateItem creates a new item in the external service from a requirement
	CreateItem(req *database.Requirement) (externalID string, err error)

	// UpdateItem updates an existing item in the external service
	UpdateItem(externalID string, req *database.Requirement) bool

	// MapStatusToRTMX maps external service status to RTMX status
	MapStatusToRTMX(externalStatus string) database.Status

	// MapStatusFromRTMX maps RTMX status to external service status
	MapStatusFromRTMX(status database.Status) string
}

// ExternalItem represents an item from an external service
type ExternalItem struct {
	ExternalID    string   // Service-specific ID (issue number, ticket key)
	Title         string   // Item title/summary
	Description   string   // Item description/body
	Status        string   // Service-specific status
	Labels        []string // Tags/labels
	URL           string   // Web URL to view item
	CreatedAt     string   // ISO timestamp
	UpdatedAt     string   // ISO timestamp
	Assignee      string   // Assigned user
	Priority      string   // Priority level
	RequirementID string   // Linked RTMX requirement ID (if found)
}

package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
)

// MondayAdapter syncs requirements with Monday.com board items via GraphQL.
type MondayAdapter struct {
	config *config.MondayAdapterConfig
	client HTTPClient
	getEnv func(string) string
	token  string
	apiURL string
}

// MondayItem represents a Monday.com board item.
type MondayItem struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	State       string              `json:"state"`
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
	Group       *MondayGroup        `json:"group"`
	ColumnValues []MondayColumnValue `json:"column_values"`
}

// MondayGroup represents a Monday.com board group.
type MondayGroup struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// MondayColumnValue represents a column value in Monday.com.
type MondayColumnValue struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
	Value string `json:"value"`
}

// NewMondayAdapter creates a new Monday.com adapter.
func NewMondayAdapter(cfg *config.MondayAdapterConfig, opts ...AdapterOption) (*MondayAdapter, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("Monday adapter is not enabled")
	}

	options := applyOptions(opts)
	tokenEnv := cfg.TokenEnv
	if tokenEnv == "" {
		tokenEnv = "MONDAY_TOKEN"
	}
	token := options.getEnv(tokenEnv)
	if token == "" {
		return nil, fmt.Errorf("Monday token not found. Set %s environment variable", tokenEnv)
	}

	return &MondayAdapter{
		config: cfg,
		client: options.httpClient,
		getEnv: options.getEnv,
		token:  token,
		apiURL: "https://api.monday.com/v2",
	}, nil
}

// SetAPIURL overrides the API URL (for testing).
func (m *MondayAdapter) SetAPIURL(url string) { m.apiURL = url }

func (m *MondayAdapter) Name() string { return "monday" }

func (m *MondayAdapter) IsConfigured() bool {
	return m.config.Enabled && m.config.BoardID != "" && m.token != ""
}

func (m *MondayAdapter) TestConnection() (bool, string) {
	query := `{ me { name } }`
	var result struct {
		Data struct {
			Me struct {
				Name string `json:"name"`
			} `json:"me"`
		} `json:"data"`
	}
	if err := m.graphQL(query, nil, &result); err != nil {
		return false, fmt.Sprintf("Connection failed: %v", err)
	}
	return true, fmt.Sprintf("Connected as %s", result.Data.Me.Name)
}

func (m *MondayAdapter) FetchItems(query map[string]interface{}) ([]ExternalItem, error) {
	gql := fmt.Sprintf(`{ boards(ids: [%s]) { items_page { items { id name state created_at updated_at group { id title } column_values { id title text } } } } }`, m.config.BoardID)
	var result struct {
		Data struct {
			Boards []struct {
				ItemsPage struct {
					Items []MondayItem `json:"items"`
				} `json:"items_page"`
			} `json:"boards"`
		} `json:"data"`
	}
	if err := m.graphQL(gql, nil, &result); err != nil {
		return nil, err
	}
	if len(result.Data.Boards) == 0 {
		return nil, nil
	}
	items := result.Data.Boards[0].ItemsPage.Items
	out := make([]ExternalItem, 0, len(items))
	for _, item := range items {
		out = append(out, m.itemToExternal(item))
	}
	return out, nil
}

func (m *MondayAdapter) GetItem(externalID string) (*ExternalItem, error) {
	gql := fmt.Sprintf(`{ items(ids: [%s]) { id name state created_at updated_at group { id title } column_values { id title text } } }`, externalID)
	var result struct {
		Data struct {
			Items []MondayItem `json:"items"`
		} `json:"data"`
	}
	if err := m.graphQL(gql, nil, &result); err != nil {
		return nil, err
	}
	if len(result.Data.Items) == 0 {
		return nil, fmt.Errorf("item %s not found", externalID)
	}
	item := m.itemToExternal(result.Data.Items[0])
	return &item, nil
}

func (m *MondayAdapter) CreateItem(req *database.Requirement) (string, error) {
	name := fmt.Sprintf("[%s] %s", req.ReqID, truncateStr(req.RequirementText, 80))
	gql := fmt.Sprintf(`mutation { create_item(board_id: %s, item_name: %q) { id } }`, m.config.BoardID, name)
	var result struct {
		Data struct {
			CreateItem struct {
				ID string `json:"id"`
			} `json:"create_item"`
		} `json:"data"`
	}
	if err := m.graphQL(gql, nil, &result); err != nil {
		return "", err
	}
	return result.Data.CreateItem.ID, nil
}

func (m *MondayAdapter) UpdateItem(externalID string, req *database.Requirement) bool {
	name := fmt.Sprintf("[%s] %s", req.ReqID, truncateStr(req.RequirementText, 80))
	gql := fmt.Sprintf(`mutation { change_simple_column_value(board_id: %s, item_id: %s, column_id: "name", value: %q) { id } }`, m.config.BoardID, externalID, name)
	var result struct {
		Data interface{} `json:"data"`
	}
	return m.graphQL(gql, nil, &result) == nil
}

func (m *MondayAdapter) MapStatusToRTMX(status string) database.Status {
	if m.config.StatusMapping != nil {
		if mapped, ok := m.config.StatusMapping[status]; ok {
			return database.Status(mapped)
		}
	}
	switch strings.ToLower(status) {
	case "done", "complete", "completed":
		return database.StatusComplete
	case "working on it", "in progress", "partial":
		return database.StatusPartial
	case "stuck", "blocked":
		return database.StatusPartial
	default:
		return database.StatusMissing
	}
}

func (m *MondayAdapter) MapStatusFromRTMX(status database.Status) string {
	switch status {
	case database.StatusComplete:
		return "Done"
	case database.StatusPartial:
		return "Working on it"
	default:
		return "Not Started"
	}
}

// GroupToCategory maps a Monday group title to an RTMX category.
func (m *MondayAdapter) GroupToCategory(groupTitle string) string {
	parts := strings.Fields(groupTitle)
	if len(parts) == 0 {
		return "UNCATEGORIZED"
	}
	return strings.ToUpper(parts[0])
}

func (m *MondayAdapter) graphQL(query string, variables map[string]interface{}, result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	payload := map[string]interface{}{"query": query}
	if variables != nil {
		payload["variables"] = variables
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", m.apiURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", m.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return fmt.Errorf("API error: HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

func (m *MondayAdapter) itemToExternal(item MondayItem) ExternalItem {
	reqID := ExtractReqID(item.Name)
	group := ""
	if item.Group != nil {
		group = item.Group.Title
	}

	// Extract status from column values
	status := item.State
	for _, col := range item.ColumnValues {
		if col.ID == "status" || col.Title == "Status" {
			if col.Text != "" {
				status = col.Text
			}
		}
	}

	return ExternalItem{
		ExternalID:    item.ID,
		Title:         item.Name,
		Status:        status,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
		Labels:        []string{group},
		RequirementID: reqID,
	}
}

package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
)

// SlackAdapter sends notifications to Slack channels via the Slack Web API.
type SlackAdapter struct {
	config *config.SlackAdapterConfig
	client HTTPClient
	getEnv func(string) string
	token  string
	apiURL string // configurable for testing; defaults to https://slack.com/api
}

// SlackAdapterOption configures optional SlackAdapter dependencies.
type SlackAdapterOption func(*SlackAdapter)

// WithSlackAPIURL sets a custom Slack API base URL (for testing).
func WithSlackAPIURL(url string) SlackAdapterOption {
	return func(s *SlackAdapter) {
		s.apiURL = url
	}
}

// NewSlackAdapter creates a new Slack adapter.
func NewSlackAdapter(cfg *config.SlackAdapterConfig, opts ...AdapterOption) (*SlackAdapter, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("slack adapter is not enabled")
	}

	options := applyOptions(opts)

	tokenEnv := cfg.TokenEnv
	if tokenEnv == "" {
		tokenEnv = "SLACK_BOT_TOKEN"
	}

	token := options.getEnv(tokenEnv)
	if token == "" {
		return nil, fmt.Errorf("slack token not found; set %s environment variable", tokenEnv)
	}

	return &SlackAdapter{
		config: cfg,
		client: options.httpClient,
		getEnv: options.getEnv,
		token:  token,
		apiURL: "https://slack.com/api",
	}, nil
}

// Name returns the adapter name.
func (s *SlackAdapter) Name() string {
	return "slack"
}

// IsConfigured checks if the adapter is properly configured.
func (s *SlackAdapter) IsConfigured() bool {
	return s.config.Enabled && s.token != "" && len(s.config.Channels) > 0
}

// SetAPIURL sets the Slack API base URL. Use this in tests.
func (s *SlackAdapter) SetAPIURL(url string) {
	s.apiURL = url
}

// slackPostMessageRequest is the payload for chat.postMessage.
type slackPostMessageRequest struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

// slackPostMessageResponse is the response from chat.postMessage.
type slackPostMessageResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// SendNotification sends a notification to the Slack channel mapped to the given event.
// If no channel is mapped for the event, it returns an error.
func (s *SlackAdapter) SendNotification(event string, message string) error {
	channel, ok := s.config.Channels[event]
	if !ok {
		return fmt.Errorf("no Slack channel configured for event %q", event)
	}

	return s.postMessage(channel, message)
}

// SendStatusUpdate sends a status summary to the "status" channel.
func (s *SlackAdapter) SendStatusUpdate(db *database.Database) error {
	channel, ok := s.config.Channels["status"]
	if !ok {
		return fmt.Errorf("no Slack channel configured for event %q", "status")
	}

	reqs := db.All()
	complete, partial, missing := 0, 0, 0
	for _, req := range reqs {
		switch {
		case req.IsComplete():
			complete++
		case req.Status == database.StatusPartial:
			partial++
		default:
			missing++
		}
	}

	total := len(reqs)
	pct := 0.0
	if total > 0 {
		pct = float64(complete) / float64(total) * 100.0
	}

	message := fmt.Sprintf("RTMX Status: %d/%d complete (%.0f%%) | %d partial | %d missing",
		complete, total, pct, partial, missing)

	return s.postMessage(channel, message)
}

// postMessage sends a chat.postMessage request to Slack.
func (s *SlackAdapter) postMessage(channel, text string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	payload := slackPostMessageRequest{
		Channel: channel,
		Text:    text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	url := s.apiURL + "/chat.postMessage"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create Slack request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("slack request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API error: HTTP %d", resp.StatusCode)
	}

	var result slackPostMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse Slack response: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("slack API error: %s", result.Error)
	}

	return nil
}

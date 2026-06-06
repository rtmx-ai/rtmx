package adapters

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rtmx-ai/rtmx/internal/config"
)

// WebhookAdapter sends outbound webhook notifications with HMAC signing and retry.
type WebhookAdapter struct {
	config *config.WebhookAdapterConfig
	client HTTPClient
	getEnv func(string) string
	secret string
	url    string
}

// NewWebhookAdapter creates a new outbound webhook adapter.
func NewWebhookAdapter(cfg *config.WebhookAdapterConfig, opts ...AdapterOption) (*WebhookAdapter, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("webhook adapter is not enabled")
	}

	if cfg.URL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}

	options := applyOptions(opts)

	secret := ""
	if cfg.SecretEnv != "" {
		secret = options.getEnv(cfg.SecretEnv)
	}

	return &WebhookAdapter{
		config: cfg,
		client: options.httpClient,
		getEnv: options.getEnv,
		secret: secret,
		url:    cfg.URL,
	}, nil
}

// Name returns the adapter name.
func (w *WebhookAdapter) Name() string {
	return "webhook"
}

// IsConfigured checks if the adapter is properly configured.
func (w *WebhookAdapter) IsConfigured() bool {
	return w.config.Enabled && w.url != ""
}

// webhookPayload is the JSON envelope sent to the webhook endpoint.
type webhookPayload struct {
	Event     string      `json:"event"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// Send sends a webhook notification for the given event.
// If the event is not in the configured events list, it is silently skipped.
// The payload is signed with HMAC-SHA256 if a secret is configured.
// Retries with exponential backoff up to MaxRetries on failure.
func (w *WebhookAdapter) Send(event string, payload interface{}) error {
	// Filter by configured events
	if len(w.config.Events) > 0 && !w.eventAllowed(event) {
		return nil // silently skip unconfigured events
	}

	envelope := webhookPayload{
		Event:     event,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      payload,
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	var lastErr error
	maxAttempts := w.config.MaxRetries + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 100ms, 200ms, 400ms, ...
			backoff := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
			time.Sleep(backoff)
		}

		lastErr = w.doSend(body)
		if lastErr == nil {
			return nil
		}
	}

	return fmt.Errorf("webhook delivery failed after %d attempts: %w", maxAttempts, lastErr)
}

// doSend performs a single webhook delivery attempt.
func (w *WebhookAdapter) doSend(body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", w.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "RTMX-Webhook/1.0")

	// Sign with HMAC-SHA256 if secret is configured
	if w.secret != "" {
		sig := computeHMACSHA256(body, []byte(w.secret))
		req.Header.Set("X-RTMX-Signature", "sha256="+sig)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook endpoint returned HTTP %d", resp.StatusCode)
	}

	return nil
}

// eventAllowed checks if the event is in the configured events list.
func (w *WebhookAdapter) eventAllowed(event string) bool {
	for _, e := range w.config.Events {
		if e == event {
			return true
		}
	}
	return false
}

// computeHMACSHA256 computes the HMAC-SHA256 signature of a message.
func computeHMACSHA256(message, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}

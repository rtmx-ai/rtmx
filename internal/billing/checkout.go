// Package billing talks to the managed Sync billing-api-v1 surface.
package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HTTPClient is the subset of http.Client used by the billing client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client calls managed Sync org and checkout endpoints.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient HTTPClient
}

func (c *Client) http() HTTPClient {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func (c *Client) base() string {
	return strings.TrimRight(c.BaseURL, "/")
}

// Org is a managed organization the caller belongs to.
type Org struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Tier string `json:"tier"`
}

// ListOrgs returns organizations for the authenticated session.
func (c *Client) ListOrgs(ctx context.Context) ([]Org, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base()+"/orgs", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	resp, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list orgs failed (%d): %s", resp.StatusCode, truncate(string(body), 200))
	}
	var parsed struct {
		Organizations []Org `json:"organizations"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	return parsed.Organizations, nil
}

// CreateOrg creates an organization owned by the caller.
func (c *Client) CreateOrg(ctx context.Context, name string) (*Org, error) {
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base()+"/orgs", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("create org failed (%d): %s", resp.StatusCode, truncate(string(body), 200))
	}
	var org Org
	if err := json.Unmarshal(body, &org); err != nil {
		return nil, err
	}
	return &org, nil
}

// CheckoutRequest creates a Stripe Checkout Session (billing-api-v1).
type CheckoutRequest struct {
	Offering   string `json:"offering"`
	Tier       string `json:"tier"`
	Quantity   int    `json:"quantity"`
	OrgID      string `json:"org_id,omitempty"`
	SuccessURL string `json:"success_url,omitempty"`
	CancelURL  string `json:"cancel_url,omitempty"`
}

// CheckoutSession is the Stripe-hosted Checkout payload.
type CheckoutSession struct {
	URL       string            `json:"url"`
	SessionID string            `json:"session_id"`
	Metadata  map[string]string `json:"metadata"`
}

// CreateCheckoutSession POSTs /billing/checkout.
func (c *Client) CreateCheckoutSession(ctx context.Context, in CheckoutRequest) (*CheckoutSession, error) {
	if in.Offering == "" {
		in.Offering = "managed_sync"
	}
	if in.Tier == "" {
		in.Tier = "team"
	}
	if in.Quantity < 1 {
		in.Quantity = 1
	}
	payload, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base()+"/billing/checkout", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("checkout failed (%d): %s", resp.StatusCode, truncate(string(body), 200))
	}
	var session CheckoutSession
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, err
	}
	if session.URL == "" {
		return nil, fmt.Errorf("checkout response missing url")
	}
	return &session, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

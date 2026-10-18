package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
)

// HTTPDoer is the subset of http.Client used for adoption heartbeats.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// AdoptionHeartbeat is the POST /adoption/heartbeat body (adoption-v1 entry fields).
type AdoptionHeartbeat struct {
	ExternalRef string `json:"external_ref"`
	OrgSlug     string `json:"org_slug"`
	Room        string `json:"room"`
	Provider    string `json:"provider,omitempty"`
	Visibility  string `json:"visibility,omitempty"`
	Source      string `json:"source,omitempty"`
}

// gitOriginURL returns the origin remote URL for the working tree at dir.
// Overridable in tests.
var gitOriginURL = func(dir string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// CanonicalGitHubRef turns a git remote URL into github.com/owner/repo.
func CanonicalGitHubRef(remote string) (string, bool) {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return "", false
	}
	remote = strings.TrimSuffix(remote, ".git")

	switch {
	case strings.HasPrefix(remote, "git@github.com:"):
		path := strings.TrimPrefix(remote, "git@github.com:")
		path = strings.Trim(path, "/")
		if path == "" || !strings.Contains(path, "/") {
			return "", false
		}
		return "github.com/" + path, true
	case strings.HasPrefix(remote, "ssh://git@github.com/"):
		path := strings.TrimPrefix(remote, "ssh://git@github.com/")
		path = strings.Trim(path, "/")
		if path == "" || !strings.Contains(path, "/") {
			return "", false
		}
		return "github.com/" + path, true
	}

	u, err := url.Parse(remote)
	if err != nil {
		return "", false
	}
	host := strings.ToLower(u.Host)
	if host == "" {
		host = strings.ToLower(u.Path)
		// bare github.com/owner/repo
		if strings.HasPrefix(host, "github.com/") {
			return strings.Trim(host, "/"), true
		}
		return "", false
	}
	if host != "github.com" && host != "www.github.com" {
		return "", false
	}
	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", false
	}
	return "github.com/" + parts[0] + "/" + parts[1], true
}

// ParseOrgRoom extracts org slug and room from a room WebSocket/HTTP URL.
// Expected path: /sync/{org}/{room}.
func ParseOrgRoom(roomURL string) (orgSlug, room string, err error) {
	normalized, err := normalizeRoomURL(roomURL)
	if err != nil {
		return "", "", err
	}
	u, err := url.Parse(normalized)
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	// sync / org / room
	if len(parts) < 3 || parts[0] != "sync" {
		return "", "", fmt.Errorf("room URL path %q does not look like /sync/{org}/{room}", u.Path)
	}
	return parts[1], parts[2], nil
}

// HTTPBaseFromRoomURL converts a ws(s) room URL to an https HTTP API base.
func HTTPBaseFromRoomURL(roomURL string) (string, error) {
	normalized, err := normalizeRoomURL(roomURL)
	if err != nil {
		return "", err
	}
	u, err := url.Parse(normalized)
	if err != nil {
		return "", err
	}
	scheme := "https"
	if u.Scheme == "ws" {
		scheme = "http"
	}
	return scheme + "://" + u.Host, nil
}

// IsManagedSyncHost reports whether host matches the managed Sync default
// or an explicit managed base (case-insensitive host compare).
func IsManagedSyncHost(roomURL, managedBase string) bool {
	base, err := HTTPBaseFromRoomURL(roomURL)
	if err != nil {
		return false
	}
	managedBase = strings.TrimRight(strings.TrimSpace(managedBase), "/")
	if managedBase == "" {
		return false
	}
	return strings.EqualFold(base, managedBase)
}

// PostAdoptionHeartbeat POSTs to {base}/adoption/heartbeat.
func PostAdoptionHeartbeat(
	ctx context.Context,
	httpClient HTTPDoer,
	baseURL, token string,
	hb AdoptionHeartbeat,
) error {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if hb.Provider == "" {
		hb.Provider = "github"
	}
	if hb.Source == "" {
		hb.Source = "cli_heartbeat"
	}
	if hb.Room == "" {
		hb.Room = "main"
	}
	payload, err := json.Marshal(hb)
	if err != nil {
		return err
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/adoption/heartbeat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("adoption heartbeat failed (%d): %s", resp.StatusCode, truncateBody(string(body), 200))
	}
	return nil
}

// BestEffortAdoptionHeartbeat resolves git origin + room path and POSTs.
// Returns nil when skipped (not managed / no git remote); never fails the caller.
func BestEffortAdoptionHeartbeat(
	ctx context.Context,
	httpClient HTTPDoer,
	roomURL, token, managedBase, workDir string,
) error {
	if token == "" || !IsManagedSyncHost(roomURL, managedBase) {
		return nil
	}
	org, room, err := ParseOrgRoom(roomURL)
	if err != nil {
		return err
	}
	remote, err := gitOriginURL(workDir)
	if err != nil {
		return fmt.Errorf("git origin: %w", err)
	}
	ref, ok := CanonicalGitHubRef(remote)
	if !ok {
		return fmt.Errorf("origin is not a GitHub remote: %s", remote)
	}
	base, err := HTTPBaseFromRoomURL(roomURL)
	if err != nil {
		return err
	}
	return PostAdoptionHeartbeat(ctx, httpClient, base, token, AdoptionHeartbeat{
		ExternalRef: ref,
		OrgSlug:     org,
		Room:        room,
	})
}

func truncateBody(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

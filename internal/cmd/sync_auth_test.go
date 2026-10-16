package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rtmx-ai/rtmx/internal/auth"
	"github.com/rtmx-ai/rtmx/internal/billing"
	syncpkg "github.com/rtmx-ai/rtmx/internal/sync"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestResolveSyncTokenPrecedence(t *testing.T) {
	rtmx.Req(t, "REQ-GO-083")

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "tokens.json")
	stored := auth.TokenSet{
		AccessToken: "stored-sess",
		Mode:        "managed",
		ExpiresAt:   time.Now().Add(time.Hour).Unix(),
	}
	data, _ := json.Marshal(stored)
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatal(err)
	}

	prevPath := authTokenPath
	authTokenPath = tokenPath
	defer func() { authTokenPath = prevPath }()

	t.Setenv("RTMX_SYNC_TOKEN", "")
	if got := resolveSyncToken(""); got != "stored-sess" {
		t.Fatalf("stored = %q", got)
	}

	t.Setenv("RTMX_SYNC_TOKEN", "env-token")
	if got := resolveSyncToken(""); got != "env-token" {
		t.Fatalf("env = %q", got)
	}

	if got := resolveSyncToken("flag-token"); got != "flag-token" {
		t.Fatalf("flag = %q", got)
	}
}

func TestResolveSyncTokenExpiredStored(t *testing.T) {
	rtmx.Req(t, "REQ-GO-083")

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "tokens.json")
	stored := auth.TokenSet{
		AccessToken: "expired-sess",
		Mode:        "managed",
		ExpiresAt:   time.Now().Add(-time.Hour).Unix(),
	}
	data, _ := json.Marshal(stored)
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatal(err)
	}

	prevPath := authTokenPath
	authTokenPath = tokenPath
	defer func() { authTokenPath = prevPath }()
	t.Setenv("RTMX_SYNC_TOKEN", "")

	if got := resolveSyncToken(""); got != "" {
		t.Fatalf("expired stored should be ignored, got %q", got)
	}
}

func TestErrMissingSyncAuthHintsLogin(t *testing.T) {
	rtmx.Req(t, "REQ-GO-083")
	err := errMissingSyncAuth()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "rtmx login") {
		t.Fatalf("expected login hint, got %q", err.Error())
	}
}

func TestPickSyncOrgSlug(t *testing.T) {
	orgs := []billing.Org{
		{ID: "1", Name: "Acme", Slug: "acme"},
		{ID: "2", Name: "Beta", Slug: "beta"},
	}
	if _, err := pickSyncOrgSlug(orgs, ""); err == nil || !strings.Contains(err.Error(), "multiple") {
		t.Fatalf("expected multiple-org error, got %v", err)
	}
	got, err := pickSyncOrgSlug(orgs, "beta")
	if err != nil || got != "beta" {
		t.Fatalf("want beta, got %q (%v)", got, err)
	}
	sole, err := pickSyncOrgSlug(orgs[:1], "")
	if err != nil || sole != "acme" {
		t.Fatalf("want acme, got %q (%v)", sole, err)
	}
}

func TestFormatRoomDialErrorUnauthenticated(t *testing.T) {
	rtmx.Req(t, "REQ-GO-083")
	err := formatRoomDialError(&syncpkg.RoomError{
		Code:   syncpkg.CloseUnauthenticated,
		Reason: "missing",
	})
	if !strings.Contains(err.Error(), "rtmx login") {
		t.Fatalf("got %v", err)
	}
}

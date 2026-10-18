package sync

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCanonicalGitHubRef(t *testing.T) {
	tests := []struct {
		in   string
		want string
		ok   bool
	}{
		{"https://github.com/rtmx-ai/rtmx.git", "github.com/rtmx-ai/rtmx", true},
		{"git@github.com:rtmx-ai/rtmx.git", "github.com/rtmx-ai/rtmx", true},
		{"ssh://git@github.com/rtmx-ai/rtmx", "github.com/rtmx-ai/rtmx", true},
		{"https://gitlab.com/acme/x", "", false},
		{"", "", false},
	}
	for _, tt := range tests {
		got, ok := CanonicalGitHubRef(tt.in)
		if ok != tt.ok || got != tt.want {
			t.Fatalf("%q: got (%q,%v) want (%q,%v)", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}

func TestParseOrgRoom(t *testing.T) {
	org, room, err := ParseOrgRoom("https://sync.rtmx.ai/sync/rtmx-ai/main")
	if err != nil {
		t.Fatal(err)
	}
	if org != "rtmx-ai" || room != "main" {
		t.Fatalf("got %s/%s", org, room)
	}
}

func TestAdoptionHeartbeat(t *testing.T) {
	t.Run("posts heartbeat", func(t *testing.T) {
		var got AdoptionHeartbeat
		var auth string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/adoption/heartbeat" {
				t.Fatalf("path %s", r.URL.Path)
			}
			auth = r.Header.Get("Authorization")
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatal(err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"external_ref":"github.com/rtmx-ai/rtmx"}`))
		}))
		defer srv.Close()

		err := PostAdoptionHeartbeat(context.Background(), srv.Client(), srv.URL, "tok", AdoptionHeartbeat{
			ExternalRef: "github.com/rtmx-ai/rtmx",
			OrgSlug:     "rtmx-ai",
			Room:        "main",
		})
		if err != nil {
			t.Fatal(err)
		}
		if auth != "Bearer tok" {
			t.Fatalf("auth %q", auth)
		}
		if got.ExternalRef != "github.com/rtmx-ai/rtmx" || got.Source != "cli_heartbeat" {
			t.Fatalf("%+v", got)
		}
	})

	t.Run("best effort skips non-managed", func(t *testing.T) {
		called := false
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()
		err := BestEffortAdoptionHeartbeat(
			context.Background(),
			srv.Client(),
			"wss://selfhost.example/sync/acme/main",
			"tok",
			"https://sync.rtmx.ai",
			".",
		)
		if err != nil {
			t.Fatal(err)
		}
		if called {
			t.Fatal("should not POST for self-hosted")
		}
	})

	t.Run("best effort posts for managed", func(t *testing.T) {
		prev := gitOriginURL
		gitOriginURL = func(string) (string, error) {
			return "https://github.com/rtmx-ai/rtmx.git", nil
		}
		defer func() { gitOriginURL = prev }()

		var path string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer srv.Close()

		roomURL := strings.Replace(srv.URL, "http://", "ws://", 1) + "/sync/rtmx-ai/main"
		err := BestEffortAdoptionHeartbeat(
			context.Background(),
			srv.Client(),
			roomURL,
			"tok",
			srv.URL,
			".",
		)
		if err != nil {
			t.Fatal(err)
		}
		if path != "/adoption/heartbeat" {
			t.Fatalf("path %q", path)
		}
	})
}

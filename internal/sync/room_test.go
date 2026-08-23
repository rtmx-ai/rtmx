package sync

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ygws "github.com/reearth/ygo/provider/websocket"

	"github.com/rtmx-ai/rtmx/internal/database"
)

// startRoomServer runs a y-websocket server so the client is exercised
// against an independent implementation of the protocol, not a mock that
// agrees with whatever the client happens to send. Convergence with the
// shipped Python server is covered by the monorepo slice test.
func startRoomServer(t *testing.T) string {
	t.Helper()

	server := httptest.NewServer(http.StripPrefix("/sync", ygws.NewServer()))
	t.Cleanup(server.Close)
	return "ws" + strings.TrimPrefix(server.URL, "http") + "/sync/acme/mvp"
}

func dial(t *testing.T, url string) *RoomClient {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	client, err := DialRoom(ctx, RoomOptions{URL: url, Timeout: 10 * time.Second})
	if err != nil {
		t.Fatalf("DialRoom: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func statusUpdate(reqID, status string) RequirementUpdate {
	return RequirementUpdate{
		ReqID:  reqID,
		Action: "updated",
		Fields: map[string]string{"status": status, "category": "DEMO"},
	}
}

// TestSyncClientRoundTrip is the acceptance test named by REQ-SYNC-002.
func TestSyncClientRoundTrip(t *testing.T) {
	url := startRoomServer(t)

	writer := dial(t, url)
	if err := writer.Publish([]RequirementUpdate{statusUpdate("REQ-DEMO-001", "COMPLETE")}); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	reader := dial(t, url)
	snapshot := reader.Snapshot()
	if len(snapshot) != 1 {
		t.Fatalf("expected 1 requirement in the room, got %d", len(snapshot))
	}
	if got := snapshot[0].Fields["status"]; got != "COMPLETE" {
		t.Errorf("status = %q, want COMPLETE", got)
	}
	if got := snapshot[0].ReqID; got != "REQ-DEMO-001" {
		t.Errorf("req id = %q, want REQ-DEMO-001", got)
	}
}

func TestRoomSnapshotAppliesToDatabase(t *testing.T) {
	url := startRoomServer(t)

	writer := dial(t, url)
	if err := writer.Publish([]RequirementUpdate{statusUpdate("REQ-DEMO-001", "COMPLETE")}); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	db := database.NewDatabase()
	req := database.NewRequirement("REQ-DEMO-001")
	req.Status = database.StatusMissing
	if err := db.Add(req); err != nil {
		t.Fatalf("Add: %v", err)
	}

	reader := dial(t, url)
	result := ApplyUpdates(db, reader.Snapshot())
	if !result.HasChanges() {
		t.Fatal("expected the room snapshot to change the database")
	}
	if got := db.Get("REQ-DEMO-001").Status; got != database.StatusComplete {
		t.Errorf("status = %q, want COMPLETE", got)
	}
}

// A second writer must not lose the first writer's work.
func TestConcurrentWritersConverge(t *testing.T) {
	url := startRoomServer(t)

	first := dial(t, url)
	second := dial(t, url)

	if err := first.Publish([]RequirementUpdate{statusUpdate("REQ-DEMO-001", "COMPLETE")}); err != nil {
		t.Fatalf("first Publish: %v", err)
	}
	if err := second.Publish([]RequirementUpdate{statusUpdate("REQ-DEMO-002", "PARTIAL")}); err != nil {
		t.Fatalf("second Publish: %v", err)
	}

	observer := dial(t, url)
	seen := make(map[string]string)
	for _, update := range observer.Snapshot() {
		seen[update.ReqID] = update.Fields["status"]
	}

	if seen["REQ-DEMO-001"] != "COMPLETE" || seen["REQ-DEMO-002"] != "PARTIAL" {
		t.Errorf("room did not converge, saw %v", seen)
	}
}

func TestRepublishIsIdempotent(t *testing.T) {
	url := startRoomServer(t)

	client := dial(t, url)
	update := statusUpdate("REQ-DEMO-001", "COMPLETE")
	for range 3 {
		if err := client.Publish([]RequirementUpdate{update}); err != nil {
			t.Fatalf("Publish: %v", err)
		}
	}

	observer := dial(t, url)
	if got := len(observer.Snapshot()); got != 1 {
		t.Errorf("republishing created %d entries, want 1", got)
	}
}

func TestDialRejectsNonRoomURLs(t *testing.T) {
	cases := map[string]string{
		"empty":       "",
		"bad scheme":  "ftp://example.com/sync/acme/mvp",
		"no room":     "ws://example.com",
		"unparseable": "ws://exa mple.com/sync",
	}

	for name, url := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			if _, err := DialRoom(ctx, RoomOptions{URL: url}); err == nil {
				t.Fatalf("expected %q to be rejected", url)
			}
		})
	}
}

func TestDialReportsUnreachableServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := DialRoom(ctx, RoomOptions{
		URL:     "ws://127.0.0.1:1/sync/acme/mvp",
		Timeout: 2 * time.Second,
	})
	if err == nil {
		t.Fatal("expected a dial error against a closed port")
	}
	if !strings.Contains(err.Error(), "could not reach") {
		t.Errorf("error should name the unreachable server, got %q", err)
	}
}

// Close codes are the server's only way to say why it refused, so they must
// reach the user as an explanation (auth-token-v1, entitlement-v1).
func TestRoomErrorExplainsCloseCodes(t *testing.T) {
	cases := map[int]string{
		CloseUnauthenticated: "not authenticated",
		CloseNotEntitled:     "entitlement",
		CloseForbidden:       "permission",
		CloseWrongOrg:        "different organization",
	}

	for code, want := range cases {
		err := &RoomError{Code: code}
		if !strings.Contains(err.Error(), want) {
			t.Errorf("close %d explained as %q, want it to mention %q", code, err, want)
		}
	}
}

func TestNormalizeRoomURLAcceptsHTTPScheme(t *testing.T) {
	got, err := normalizeRoomURL("https://sync.example.com/sync/acme/mvp")
	if err != nil {
		t.Fatalf("normalizeRoomURL: %v", err)
	}
	if got != "wss://sync.example.com/sync/acme/mvp" {
		t.Errorf("got %q, want the wss form", got)
	}
}

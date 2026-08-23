package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	ygws "github.com/reearth/ygo/provider/websocket"

	"github.com/rtmx-ai/rtmx/internal/database"
	syncpkg "github.com/rtmx-ai/rtmx/internal/sync"
)

// TestServeFollowsRoomIntoDatabase covers REQ-SYNC-002 AC4: with --sync-url a
// served dashboard reflects what collaborators write, rather than only
// printing the URL it was given.
func TestServeFollowsRoomIntoDatabase(t *testing.T) {
	server := httptest.NewServer(http.StripPrefix("/sync", ygws.NewServer()))
	defer server.Close()
	roomURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/sync/acme/mvp"

	db := database.NewDatabase()
	req := database.NewRequirement("REQ-DEMO-001")
	req.Category = "DEMO"
	req.Status = database.StatusMissing
	if err := db.Add(req); err != nil {
		t.Fatalf("Add: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "database.csv")
	if err := db.Save(dbPath); err != nil {
		t.Fatalf("Save: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lock := &sync.Mutex{}
	go func() {
		_ = followRoomIntoDatabase(ctx, roomFollower{
			db:       db,
			dbPath:   dbPath,
			lock:     lock,
			url:      roomURL,
			interval: 20 * time.Millisecond,
		})
	}()

	writer, err := syncpkg.DialRoom(ctx, syncpkg.RoomOptions{URL: roomURL, Timeout: 10 * time.Second})
	if err != nil {
		t.Fatalf("DialRoom: %v", err)
	}
	defer func() { _ = writer.Close() }()

	err = writer.Publish([]syncpkg.RequirementUpdate{{
		ReqID:  "REQ-DEMO-001",
		Action: "updated",
		Fields: map[string]string{"status": "COMPLETE", "category": "DEMO"},
	}})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		lock.Lock()
		status := db.Get("REQ-DEMO-001").Status
		lock.Unlock()
		if status == database.StatusComplete {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("dashboard database still shows %q, want COMPLETE", status)
		}
		time.Sleep(10 * time.Millisecond)
	}

	saved, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(saved), "COMPLETE") {
		t.Error("remote write was not persisted to the CSV")
	}
}

// A dashboard is useful against the local CSV even when the sync server is
// unreachable, so serve must not die on a refused connection.
func TestServeSurvivesAnUnreachableSyncServer(t *testing.T) {
	db := database.NewDatabase()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reported int
	done := make(chan error, 1)
	go func() {
		done <- followRoomIntoDatabase(ctx, roomFollower{
			db:       db,
			lock:     &sync.Mutex{},
			url:      "ws://127.0.0.1:1/sync/acme/mvp",
			interval: 20 * time.Millisecond,
			logf: func(string, ...any) {
				reported++
			},
		})
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("follower returned %v, want it to stop quietly on cancel", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("follower did not stop when its context was cancelled")
	}
	if reported == 0 {
		t.Error("an unreachable sync server should be reported to the user")
	}
}

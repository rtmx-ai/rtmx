package sync

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// waitFor polls until cond holds, so a follower running on its own schedule
// does not need a sleep long enough to be flaky.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

// A dashboard started with --sync-url must show what other people write
// (REQ-SYNC-002 AC4).
func TestFollowDeliversRemoteWrites(t *testing.T) {
	url := startRoomServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	seen := make(map[string]string)

	go func() {
		_ = Follow(ctx, FollowOptions{
			RoomOptions: RoomOptions{URL: url, Timeout: 5 * time.Second},
			Interval:    20 * time.Millisecond,
			OnSnapshot: func(updates []RequirementUpdate) {
				mu.Lock()
				defer mu.Unlock()
				for _, update := range updates {
					seen[update.ReqID] = update.Fields["status"]
				}
			},
		})
	}()

	writer := dial(t, url)
	if err := writer.Publish([]RequirementUpdate{statusUpdate("REQ-DEMO-001", "COMPLETE")}); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	waitFor(t, "the follower to observe the remote write", func() bool {
		mu.Lock()
		defer mu.Unlock()
		return seen["REQ-DEMO-001"] == "COMPLETE"
	})
}

func TestFollowStopsOnContextCancel(t *testing.T) {
	url := startRoomServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- Follow(ctx, FollowOptions{
			RoomOptions: RoomOptions{URL: url, Timeout: 5 * time.Second},
			Interval:    20 * time.Millisecond,
		})
	}()

	cancel()
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("Follow returned %v, want nil or context.Canceled", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Follow ignored context cancellation")
	}
}

// Retrying a refusal the user must act on (bad token, no entitlement) would
// bury the reason under reconnect noise.
func TestFollowDoesNotRetryARefusal(t *testing.T) {
	url := startRoomServer(t)

	attempts := 0
	err := Follow(context.Background(), FollowOptions{
		RoomOptions: RoomOptions{URL: url, Timeout: time.Second},
		Interval:    10 * time.Millisecond,
		dialRoom: func(context.Context, RoomOptions) (*RoomClient, error) {
			attempts++
			return nil, &RoomError{Code: CloseNotEntitled, Reason: "trial expired"}
		},
	})

	var roomErr *RoomError
	if !errors.As(err, &roomErr) || roomErr.Code != CloseNotEntitled {
		t.Fatalf("Follow returned %v, want the 4402 refusal", err)
	}
	if attempts != 1 {
		t.Errorf("dialed %d times, want 1: a refusal is not retryable", attempts)
	}
}

// A server that is merely down should not kill a dashboard that is otherwise
// usable against the local CSV.
func TestFollowRetriesATransientDialFailure(t *testing.T) {
	url := startRoomServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	attempts := 0
	failures := 0

	go func() {
		_ = Follow(ctx, FollowOptions{
			RoomOptions: RoomOptions{URL: url, Timeout: 5 * time.Second},
			Interval:    20 * time.Millisecond,
			OnError: func(error) {
				mu.Lock()
				defer mu.Unlock()
				failures++
			},
			dialRoom: func(ctx context.Context, opts RoomOptions) (*RoomClient, error) {
				mu.Lock()
				attempts++
				first := attempts == 1
				mu.Unlock()
				if first {
					return nil, errors.New("connection refused")
				}
				return DialRoom(ctx, opts)
			},
		})
	}()

	waitFor(t, "the follower to reconnect after a transient failure", func() bool {
		mu.Lock()
		defer mu.Unlock()
		return attempts > 1 && failures > 0
	})
}

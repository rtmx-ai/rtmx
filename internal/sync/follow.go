package sync

import (
	"context"
	"errors"
	"time"
)

// DefaultFollowInterval is how often a follower refreshes room state.
const DefaultFollowInterval = 5 * time.Second

// followRetryDelay paces reconnection after the room drops. It is short
// enough that a restarted server is picked up while a user is still looking
// at the dashboard.
const followRetryDelay = 2 * time.Second

// FollowOptions configures a long-lived room subscription.
type FollowOptions struct {
	RoomOptions

	// Interval is the refresh period. Defaults to DefaultFollowInterval.
	Interval time.Duration

	// OnSnapshot receives the room's requirements after every refresh,
	// including the first one taken at connect.
	OnSnapshot func([]RequirementUpdate)

	// OnError reports recoverable trouble (a dropped connection, a server
	// restart). Following continues afterwards.
	OnError func(error)

	// dialRoom exists so tests can inject connection failures.
	dialRoom func(context.Context, RoomOptions) (*RoomClient, error)
}

// Follow keeps a room connection open and reports its state until ctx is
// cancelled (REQ-SYNC-002 AC4).
//
// Transient trouble is retried, but a refusal the user must act on -- a bad
// token, a missing entitlement, the wrong organization -- is returned instead,
// because reconnecting cannot fix it and would bury the reason.
func Follow(ctx context.Context, opts FollowOptions) error {
	interval := opts.Interval
	if interval <= 0 {
		interval = DefaultFollowInterval
	}
	dial := opts.dialRoom
	if dial == nil {
		dial = DialRoom
	}

	for {
		if err := ctx.Err(); err != nil {
			return nil
		}

		client, err := dial(ctx, opts.RoomOptions)
		if err != nil {
			if isRefusal(err) {
				return err
			}
			report(opts.OnError, err)
			if !sleepUntil(ctx, followRetryDelay) {
				return nil
			}
			continue
		}

		err = followRoom(ctx, client, interval, opts.OnSnapshot)
		_ = client.Close()

		switch {
		case err == nil:
			return nil
		case isRefusal(err):
			return err
		default:
			report(opts.OnError, err)
			if !sleepUntil(ctx, followRetryDelay) {
				return nil
			}
		}
	}
}

// followRoom refreshes one connection until it fails or ctx ends. A nil error
// means the caller asked to stop.
func followRoom(
	ctx context.Context,
	client *RoomClient,
	interval time.Duration,
	onSnapshot func([]RequirementUpdate),
) error {
	deliver := func() {
		if onSnapshot != nil {
			onSnapshot(client.Snapshot())
		}
	}
	deliver()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := client.Refresh(); err != nil {
				return err
			}
			deliver()
		}
	}
}

func isRefusal(err error) bool {
	var roomErr *RoomError
	return errors.As(err, &roomErr)
}

func report(onError func(error), err error) {
	if onError != nil {
		onError(err)
	}
}

// sleepUntil reports whether the wait finished rather than being cancelled.
func sleepUntil(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

package cmd

import (
	"context"
	"sync"
	"time"

	"github.com/rtmx-ai/rtmx/internal/database"
	syncpkg "github.com/rtmx-ai/rtmx/internal/sync"
)

// roomFollower feeds a served dashboard from an rtmx-sync room
// (REQ-SYNC-002 AC4).
type roomFollower struct {
	db     *database.Database
	dbPath string
	lock   *sync.Mutex
	url    string
	token  string

	// interval overrides the refresh period; zero uses the package default.
	interval time.Duration

	logf func(format string, args ...any)
}

// followRoomIntoDatabase applies remote room state to the served database
// until ctx is cancelled.
//
// Remote writes are persisted to the CSV for the same reason dashboard edits
// are: the file, not the process, is what the user's other tools read.
func followRoomIntoDatabase(ctx context.Context, f roomFollower) error {
	return syncpkg.Follow(ctx, syncpkg.FollowOptions{
		RoomOptions: syncpkg.RoomOptions{URL: f.url, Token: f.token},
		Interval:    f.interval,
		OnSnapshot: func(updates []syncpkg.RequirementUpdate) {
			f.apply(updates)
		},
		OnError: func(err error) {
			f.log("  Sync interrupted, retrying: %v\n", err)
		},
	})
}

func (f roomFollower) apply(updates []syncpkg.RequirementUpdate) {
	if len(updates) == 0 {
		return
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	result := syncpkg.ApplyUpdates(f.db, updates)
	if !result.HasChanges() {
		return
	}
	if f.dbPath != "" {
		if err := f.db.Save(f.dbPath); err != nil {
			f.log("  Could not write %s: %v\n", f.dbPath, err)
			return
		}
	}
	f.log("  Sync: %s\n", result.Summary())
}

func (f roomFollower) log(format string, args ...any) {
	if f.logf != nil {
		f.logf(format, args...)
	}
}

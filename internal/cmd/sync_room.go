package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rtmx-ai/rtmx/internal/auth"
	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/output"
	syncpkg "github.com/rtmx-ai/rtmx/internal/sync"
)

// runRoomSync joins an rtmx-sync room, exchanges state, and writes the
// merged result back to the CSV (REQ-SYNC-002).
//
// The CSV stays the source of truth on disk: the room is a shared view of it,
// so every path through this function ends by saving what the room agreed on.
func runRoomSync(cfg *config.Config) error {
	if syncPull && syncPush {
		return NewExitError(1, "use --pull or --push, not both")
	}
	if !syncPull && !syncPush && len(syncSet) == 0 {
		return NewExitError(1, "specify --pull, --push, or --set REQ-ID=STATUS")
	}

	dbPath := cfg.DatabasePath(".")
	if dbPath == "" {
		dbPath = ".rtmx/database.csv"
	}
	db, err := loadOrCreateDatabase(dbPath)
	if err != nil {
		return NewExitError(1, err.Error())
	}

	setUpdates, err := parseSetUpdates(syncSet, db)
	if err != nil {
		return NewExitError(1, err.Error())
	}

	fmt.Printf("%sConnecting to %s...%s\n", output.Bold, syncURL, output.Reset)
	token := roomToken()
	if token == "" {
		fmt.Printf("  %s✗%s %v\n", output.Red, output.Reset, errMissingSyncAuth())
		return errMissingSyncAuth()
	}
	ctx, cancel := context.WithTimeout(context.Background(), roomTimeout())
	defer cancel()

	client, err := syncpkg.DialRoom(ctx, syncpkg.RoomOptions{
		URL:     syncURL,
		Token:   token,
		Timeout: roomTimeout(),
	})
	if err != nil {
		fmt.Printf("  %s✗%s %v\n", output.Red, output.Reset, formatRoomDialError(err))
		return NewExitError(1, formatRoomDialError(err).Error())
	}
	defer func() { _ = client.Close() }()
	fmt.Printf("  %s✓%s joined room\n\n", output.Green, output.Reset)

	outbound := setUpdates
	if syncPush {
		outbound = syncpkg.DatabaseToUpdates(db)
	}
	if len(outbound) > 0 {
		if err := client.Publish(outbound); err != nil {
			fmt.Printf("  %s✗%s %v\n", output.Red, output.Reset, err)
			return NewExitError(1, err.Error())
		}
		fmt.Printf("Pushed %d requirement(s) to the room\n", len(outbound))
	}

	result := syncpkg.ApplyUpdates(db, client.Snapshot())
	if err := db.Save(dbPath); err != nil {
		return NewExitError(1, fmt.Sprintf("could not write %s: %v", dbPath, err))
	}

	fmt.Printf("Local database: %s\n", result.Summary())
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			fmt.Printf("  %s✗%s %s\n", output.Red, output.Reset, e)
		}
		return NewExitError(1, "room sync completed with errors")
	}
	return nil
}

// loadOrCreateDatabase tolerates a first sync into an empty workspace.
func loadOrCreateDatabase(path string) (*database.Database, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return database.NewDatabase(), nil
		}
		return nil, err
	}
	return database.Load(path)
}

// parseSetUpdates turns --set REQ-ID=STATUS into room updates, carrying the
// requirement's other fields so a new room is not populated with blanks.
func parseSetUpdates(args []string, db *database.Database) ([]syncpkg.RequirementUpdate, error) {
	if len(args) == 0 {
		return nil, nil
	}

	byID := make(map[string]syncpkg.RequirementUpdate)
	for _, update := range syncpkg.DatabaseToUpdates(db) {
		byID[update.ReqID] = update
	}

	updates := make([]syncpkg.RequirementUpdate, 0, len(args))
	for _, arg := range args {
		reqID, value, found := strings.Cut(arg, "=")
		reqID = strings.TrimSpace(reqID)
		value = strings.TrimSpace(value)
		if !found || reqID == "" || value == "" {
			return nil, fmt.Errorf("invalid --set %q, expected REQ-ID=STATUS", arg)
		}
		status, err := database.ParseStatus(value)
		if err != nil {
			return nil, fmt.Errorf("invalid status in --set %q: %w", arg, err)
		}

		update, known := byID[reqID]
		if !known {
			update = syncpkg.RequirementUpdate{
				ReqID:  reqID,
				Action: "added",
				Fields: map[string]string{},
			}
		}
		update.Fields["status"] = string(status)
		update.Timestamp = time.Now()
		updates = append(updates, update)
	}
	return updates, nil
}

func roomToken() string {
	return resolveSyncToken(syncToken)
}

// resolveSyncToken returns the room credential with precedence:
// explicit flag/value, RTMX_SYNC_TOKEN, then stored managed login (REQ-GO-083).
func resolveSyncToken(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if v := os.Getenv("RTMX_SYNC_TOKEN"); v != "" {
		return v
	}
	tokens, err := auth.LoadStoredTokens(resolveTokenPath())
	if err != nil || tokens == nil || tokens.AccessToken == "" {
		return ""
	}
	if tokens.IsExpired() {
		return ""
	}
	return tokens.AccessToken
}

// errMissingSyncAuth is returned when room sync has no usable credential.
func errMissingSyncAuth() error {
	return NewExitError(1, "not authenticated: run 'rtmx login' or pass --token / RTMX_SYNC_TOKEN")
}


func roomTimeout() time.Duration {
	if syncTimeout > 0 {
		return time.Duration(syncTimeout) * time.Second
	}
	return 30 * time.Second
}

// formatRoomDialError maps server close codes to actionable CLI guidance.
func formatRoomDialError(err error) error {
	var roomErr *syncpkg.RoomError
	if errors.As(err, &roomErr) && roomErr.Code == syncpkg.CloseUnauthenticated {
		return errMissingSyncAuth()
	}
	return err
}


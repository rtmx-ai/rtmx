// Package sync implements RTM synchronization with external services and
// with an rtmx-sync collaboration room.
package sync

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/reearth/ygo/crdt"
	ysync "github.com/reearth/ygo/sync"
)

// RequirementsMapName is the root Y.Map holding requirement data, per
// system/contracts/sync-protocol-v1.json.
const RequirementsMapName = "requirements"

// messageSync is the y-websocket outer frame type that wraps a y-protocols
// sync message. Awareness (1) is not used by the CLI.
const messageSync = 0

// Close codes the server uses to explain a refusal (auth-token-v1,
// entitlement-v1). They are reported to the user rather than surfacing as a
// bare "websocket closed" so a failed sync is actionable.
const (
	CloseUnauthenticated = 4401
	CloseNotEntitled     = 4402
	CloseForbidden       = 4403
	CloseWrongOrg        = 4409
)

type localOrigin struct{}
type remoteOrigin struct{}

// RoomError is a refusal the server explained with a close code.
type RoomError struct {
	Code   int
	Reason string
}

func (e *RoomError) Error() string {
	switch e.Code {
	case CloseUnauthenticated:
		return "not authenticated: pass --token or set RTMX_SYNC_TOKEN"
	case CloseNotEntitled:
		return fmt.Sprintf("no active entitlement for this organization (%s)", e.Reason)
	case CloseForbidden:
		return "credential lacks permission for this room"
	case CloseWrongOrg:
		return "credential belongs to a different organization than the room"
	}
	if e.Reason != "" {
		return fmt.Sprintf("sync server closed the connection (%d): %s", e.Code, e.Reason)
	}
	return fmt.Sprintf("sync server closed the connection (%d)", e.Code)
}

// RoomOptions configures a room connection.
type RoomOptions struct {
	// URL is the room address, e.g. ws://host:1234/sync/acme/mvp.
	URL string
	// Token authenticates the connection: an API key or session token.
	Token string
	// Timeout bounds the whole conversation, not just the dial.
	Timeout time.Duration
}

// RoomClient is a y-websocket client for one rtmx-sync room.
//
// It is deliberately one-shot and synchronous: the CLI connects, exchanges
// state, writes the CSV, and exits. Long-lived presence belongs to
// `rtmx serve`, not to a command a script runs in a loop.
type RoomClient struct {
	conn    *websocket.Conn
	doc     *crdt.Doc
	pending [][]byte
	unsub   func()
	timeout time.Duration
}

// DialRoom connects to a room and completes the sync-protocol-v1 handshake.
func DialRoom(ctx context.Context, opts RoomOptions) (*RoomClient, error) {
	endpoint, err := normalizeRoomURL(opts.URL)
	if err != nil {
		return nil, err
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	header := http.Header{}
	if opts.Token != "" {
		header.Set("Authorization", "Bearer "+opts.Token)
	}

	dialer := &websocket.Dialer{HandshakeTimeout: timeout}
	conn, resp, err := dialer.DialContext(ctx, endpoint, header)
	if err != nil {
		return nil, dialError(endpoint, resp, err)
	}

	client := &RoomClient{
		conn:    conn,
		doc:     crdt.New(),
		timeout: timeout,
	}
	client.unsub = client.doc.OnUpdate(func(update []byte, origin any) {
		if _, remote := origin.(remoteOrigin); remote {
			return
		}
		client.pending = append(client.pending, update)
	})

	if err := client.handshake(); err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}

// Close releases the connection.
func (c *RoomClient) Close() error {
	if c.unsub != nil {
		c.unsub()
		c.unsub = nil
	}
	if c.conn == nil {
		return nil
	}
	_ = c.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(time.Second),
	)
	err := c.conn.Close()
	c.conn = nil
	return err
}

// Snapshot returns the room's requirements as updates ready for ApplyUpdates.
func (c *RoomClient) Snapshot() []RequirementUpdate {
	root := c.doc.GetMap(RequirementsMapName)
	now := time.Now()

	var updates []RequirementUpdate
	for _, reqID := range root.Keys() {
		value, ok := root.Get(reqID)
		if !ok {
			continue
		}
		fields := decodeRequirement(value)
		if len(fields) == 0 {
			continue
		}
		updates = append(updates, RequirementUpdate{
			ReqID:     reqID,
			Action:    "updated",
			Fields:    fields,
			Source:    "room",
			Timestamp: now,
		})
	}
	return updates
}

// Publish writes updates into the room and waits for the server to process
// them.
//
// The wait matters for a one-shot command: without it the process can exit
// with bytes still in flight and the user is told a push succeeded that the
// server never saw.
func (c *RoomClient) Publish(updates []RequirementUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	root := c.doc.GetMap(RequirementsMapName)
	c.doc.Transact(func(txn *crdt.Transaction) {
		for _, update := range updates {
			if update.Action == "removed" {
				root.Delete(txn, update.ReqID)
				continue
			}
			entry := crdt.NewMapPrelim()
			for key, value := range update.Fields {
				entry.Set(txn, key, value)
			}
			root.Set(txn, update.ReqID, entry)
		}
	}, localOrigin{})

	for _, update := range c.pending {
		if err := c.send(ysync.EncodeUpdate(update)); err != nil {
			return err
		}
	}
	c.pending = nil

	// The server applies messages in order, so its answer to a fresh
	// SyncStep1 proves the updates above were processed.
	return c.roundTrip()
}

// Refresh pulls whatever the room has learned since the last exchange.
func (c *RoomClient) Refresh() error {
	return c.roundTrip()
}

func (c *RoomClient) handshake() error {
	if err := c.send(ysync.EncodeSyncStep1(c.doc)); err != nil {
		return err
	}
	return c.readUntilSynced()
}

// roundTrip asks for state again and waits for the reply.
func (c *RoomClient) roundTrip() error {
	if err := c.send(ysync.EncodeSyncStep1(c.doc)); err != nil {
		return err
	}
	return c.readUntilSynced()
}

// readUntilSynced consumes frames until the server answers our SyncStep1.
func (c *RoomClient) readUntilSynced() error {
	deadline := time.Now().Add(c.timeout)
	if err := c.conn.SetReadDeadline(deadline); err != nil {
		return err
	}

	for {
		_, frame, err := c.conn.ReadMessage()
		if err != nil {
			return readError(err)
		}
		if len(frame) == 0 {
			continue
		}
		if frame[0] != messageSync {
			continue // awareness and anything else the CLI does not model
		}

		message := frame[1:]
		msgType, _, err := ysync.ReadSyncMessage(message)
		if err != nil {
			return fmt.Errorf("malformed sync message from server: %w", err)
		}

		reply, err := ysync.ApplySyncMessage(c.doc, message, remoteOrigin{})
		if err != nil {
			return fmt.Errorf("could not apply server state: %w", err)
		}
		if reply != nil {
			if err := c.send(reply); err != nil {
				return err
			}
		}
		if msgType == ysync.MsgSyncStep2 {
			return nil
		}
	}
}

func (c *RoomClient) send(message []byte) error {
	frame := make([]byte, 0, len(message)+1)
	frame = append(frame, messageSync)
	frame = append(frame, message...)

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return err
	}
	if err := c.conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		return readError(err)
	}
	return nil
}

// decodeRequirement reads one requirement entry, which is a nested Y.Map of
// scalar fields.
func decodeRequirement(value any) map[string]string {
	entry, ok := value.(*crdt.YMap)
	if !ok {
		return nil
	}
	fields := make(map[string]string)
	for key, raw := range entry.Entries() {
		if text, ok := raw.(string); ok {
			fields[key] = text
		}
	}
	return fields
}

// normalizeRoomURL accepts http(s) as a convenience and requires a room path.
func normalizeRoomURL(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", errors.New("sync URL is empty")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid sync URL %q: %w", raw, err)
	}

	switch parsed.Scheme {
	case "ws", "wss":
	case "http":
		parsed.Scheme = "ws"
	case "https":
		parsed.Scheme = "wss"
	default:
		return "", fmt.Errorf("sync URL must be ws:// or wss://, got %q", parsed.Scheme)
	}

	if strings.Trim(parsed.Path, "/") == "" {
		return "", fmt.Errorf("sync URL %q has no room path, expected /sync/{org}/{room}", raw)
	}
	return parsed.String(), nil
}

// dialError turns a refused upgrade into something a user can act on.
func dialError(endpoint string, resp *http.Response, err error) error {
	if resp == nil {
		return fmt.Errorf("could not reach %s: %w", endpoint, err)
	}
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return &RoomError{Code: CloseUnauthenticated}
	case http.StatusForbidden:
		// A server that refuses during the handshake cannot send a close
		// code, so the specific reason is not knowable from here.
		return fmt.Errorf(
			"%s refused the connection (HTTP 403): the credential may be missing, "+
				"scoped to another organization, or the room may be closed to it", endpoint)
	case http.StatusNotFound:
		return fmt.Errorf("%s has no such room", endpoint)
	}
	return fmt.Errorf("could not open %s (HTTP %d): %w", endpoint, resp.StatusCode, err)
}

// readError maps a close frame to a RoomError so callers can explain it.
func readError(err error) error {
	var closeErr *websocket.CloseError
	if errors.As(err, &closeErr) {
		if closeErr.Code == websocket.CloseNormalClosure {
			return err
		}
		return &RoomError{Code: closeErr.Code, Reason: closeErr.Text}
	}
	return err
}

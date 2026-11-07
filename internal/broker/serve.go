package broker

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/rtmx-ai/rtmx/internal/attestation"
)

// WireRequest is the JSON body a client sends over the broker socket.
type WireRequest struct {
	RequirementID string     `json:"requirement_id"`
	CommitSHA     string     `json:"commit_sha"`
	ClientCommand []string   `json:"client_command,omitempty"`
	Gate          GateResult `json:"gate"`
}

// WireResponse is the JSON body the broker returns.
type WireResponse struct {
	OK          bool                `json:"ok"`
	Error       string              `json:"error,omitempty"`
	Attestation *attestation.Record `json:"attestation,omitempty"`
	Output      []byte              `json:"output,omitempty"`
}

// ServeConfig configures the packaged verifier daemon.
type ServeConfig struct {
	SocketPath string
	KeyPath    string
	WorkDir    string
	Command    []string
	Key        ed25519.PrivateKey // optional; loaded from KeyPath when nil
}

// LoadKey reads an ed25519 private key (64 bytes) from path.
func LoadKey(path string) (ed25519.PrivateKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("broker: key %s has %d bytes, want %d", path, len(raw), ed25519.PrivateKeySize)
	}
	return ed25519.PrivateKey(raw), nil
}

// Serve listens on SocketPath and handles one JSON request per connection.
func Serve(ctx context.Context, cfg ServeConfig) error {
	if cfg.SocketPath == "" {
		cfg.SocketPath = DefaultBrokerSocket()
	}
	key := cfg.Key
	if key == nil {
		var err error
		key, err = LoadKey(cfg.KeyPath)
		if err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(cfg.SocketPath), 0o755); err != nil {
		return err
	}
	_ = os.Remove(cfg.SocketPath)
	ln, err := net.Listen("unix", cfg.SocketPath)
	if err != nil {
		return fmt.Errorf("broker: listen %s: %w", cfg.SocketPath, err)
	}
	defer func() { _ = ln.Close() }()
	_ = os.Chmod(cfg.SocketPath, 0o660)

	errCh := make(chan error, 1)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					errCh <- nil
				default:
					errCh <- err
				}
				return
			}
			go handleConn(conn, Config{
				WorkDir: cfg.WorkDir,
				Command: cfg.Command,
				Key:     key,
			})
		}
	}()

	select {
	case <-ctx.Done():
		_ = ln.Close()
		return nil
	case err := <-errCh:
		return err
	}
}

func handleConn(conn net.Conn, cfg Config) {
	defer func() { _ = conn.Close() }()
	var req WireRequest
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		_ = json.NewEncoder(conn).Encode(WireResponse{OK: false, Error: err.Error()})
		return
	}
	res, err := cfg.Run(Request(req))
	if err != nil {
		_ = json.NewEncoder(conn).Encode(WireResponse{OK: false, Error: err.Error()})
		return
	}
	_ = json.NewEncoder(conn).Encode(WireResponse{
		OK:          true,
		Attestation: res.Attestation,
		Output:      res.Output,
	})
}

// DialSubmit sends a verify request to a running broker socket.
func DialSubmit(socketPath string, req Request) (*Result, error) {
	if socketPath == "" {
		socketPath = DefaultBrokerSocket()
	}
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("verify: broker not reachable at %s (%v); run `sudo rtmx broker install`", socketPath, err)
	}
	defer func() { _ = conn.Close() }()
	if err := json.NewEncoder(conn).Encode(WireRequest(req)); err != nil {
		return nil, err
	}
	var resp WireResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("%s", resp.Error)
	}
	return &Result{Attestation: resp.Attestation, Output: resp.Output}, nil
}

// SocketReachable reports whether a broker is listening (no admin).
func SocketReachable(socketPath string) bool {
	if socketPath == "" {
		socketPath = DefaultBrokerSocket()
	}
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

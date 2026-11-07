package broker

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestServeAndDialSubmitRoundTrip(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix socket protocol")
	}
	dir, sha := gitRepo(t)
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	// macOS sun_path is short; keep the socket path under the limit.
	sock := filepath.Join(os.TempDir(), fmt.Sprintf("rtmx-verify-%d.sock", os.Getpid()))
	defer func() { _ = os.Remove(sock) }()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- Serve(ctx, ServeConfig{
			SocketPath: sock,
			WorkDir:    dir,
			Command:    okCommand(),
			Key:        priv,
		})
	}()
	deadline := time.Now().Add(5 * time.Second)
	for !SocketReachable(sock) {
		select {
		case err := <-serveErr:
			t.Fatalf("serve exited early: %v", err)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("broker socket never became reachable")
		}
		time.Sleep(20 * time.Millisecond)
	}
	res, err := DialSubmit(sock, Request{
		RequirementID: "REQ-MONO-069d",
		CommitSHA:     sha,
		Gate:          GateResult{TestsPassed: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Attestation == nil || res.Attestation.CommitSHA != sha {
		t.Fatalf("attestation = %+v", res.Attestation)
	}
}

func TestDialSubmitHintsBrokerInstallWhenMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix socket protocol")
	}
	_, err := DialSubmit(filepath.Join(t.TempDir(), "missing.sock"), Request{
		RequirementID: "REQ-X",
		CommitSHA:     "abc",
		Gate:          GateResult{TestsPassed: 1},
	})
	if err == nil || !strings.Contains(err.Error(), "broker install") {
		t.Fatalf("err = %v", err)
	}
}

func TestLoadKeyRoundTrip(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "verify.key")
	if err := os.WriteFile(path, priv, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := LoadKey(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(priv) {
		t.Fatalf("key len = %d", len(got))
	}
}

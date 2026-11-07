package broker

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestEnsureBrokerInstalledSkipsWhenReachable(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-069j")
	elevated := false
	err := EnsureBrokerInstalled(EnsureOptions{
		SocketPath: "dummy.sock",
		MarkerDir:  t.TempDir(),
		Reachable:  func(string) bool { return true },
		Elevate: func([]string) error {
			elevated = true
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if elevated {
		t.Fatal("must not elevate when broker is reachable")
	}
}

func TestEnsureBrokerInstalledElevatesOnceWhenMissing(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-069j")
	marker := t.TempDir()
	calls := 0
	reachable := false
	err := EnsureBrokerInstalled(EnsureOptions{
		SocketPath: "missing.sock",
		MarkerDir:  marker,
		Reachable:  func(string) bool { return reachable },
		Elevate: func(argv []string) error {
			calls++
			if len(argv) < 3 || argv[1] != "broker" || argv[2] != "install" {
				t.Fatalf("argv = %v", argv)
			}
			reachable = true
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("elevate calls = %d", calls)
	}
}

func TestEnsureBrokerInstalledRespectsCooldown(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-069j")
	dir := t.TempDir()
	marker := AttemptMarkerPath(dir)
	if err := os.WriteFile(marker, []byte(time.Now().UTC().Format(time.RFC3339)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	elevated := false
	err := EnsureBrokerInstalled(EnsureOptions{
		SocketPath: "missing.sock",
		MarkerDir:  dir,
		Reachable:  func(string) bool { return false },
		Elevate: func([]string) error {
			elevated = true
			return nil
		},
	})
	if err == nil {
		t.Fatal("expected cooldown error")
	}
	if elevated {
		t.Fatal("must not elevate during cooldown")
	}
}

func TestEnsureBrokerInstalledSkipElevate(t *testing.T) {
	err := EnsureBrokerInstalled(EnsureOptions{
		SocketPath:  "missing.sock",
		MarkerDir:   t.TempDir(),
		SkipElevate: true,
		Reachable:   func(string) bool { return false },
		Elevate: func([]string) error {
			t.Fatal("elevate must not run")
			return nil
		},
	})
	if err == nil {
		t.Fatal("expected skip error")
	}
}

func TestEnsureBrokerInstalledPreferUpgrade(t *testing.T) {
	reachable := false
	err := EnsureBrokerInstalled(EnsureOptions{
		SocketPath:    "missing.sock",
		MarkerDir:     t.TempDir(),
		PreferUpgrade: true,
		Reachable:     func(string) bool { return reachable },
		Elevate: func(argv []string) error {
			if argv[2] != "upgrade" {
				t.Fatalf("argv = %v", argv)
			}
			reachable = true
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEnsureBrokerInstalledElevateFailure(t *testing.T) {
	dir := t.TempDir()
	err := EnsureBrokerInstalled(EnsureOptions{
		SocketPath: "missing.sock",
		MarkerDir:  dir,
		Reachable:  func(string) bool { return false },
		Elevate:    func([]string) error { return errors.New("denied") },
	})
	if err == nil {
		t.Fatal("expected failure")
	}
	if _, statErr := os.Stat(filepath.Join(dir, "ensure-broker.attempt")); statErr != nil {
		t.Fatalf("marker missing: %v", statErr)
	}
}

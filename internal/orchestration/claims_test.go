package orchestration

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestClaimProtocol(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-005")

	t.Run("claim_and_release", func(t *testing.T) {
		store := newTestStore(t)

		claim, err := store.Claim("REQ-001", "agent-1")
		if err != nil {
			t.Fatalf("Claim failed: %v", err)
		}
		if claim.ReqID != "REQ-001" {
			t.Errorf("ReqID = %q, want REQ-001", claim.ReqID)
		}
		if claim.AgentID != "agent-1" {
			t.Errorf("AgentID = %q, want agent-1", claim.AgentID)
		}
		if claim.ClaimedAt.IsZero() {
			t.Error("ClaimedAt should not be zero")
		}

		// Verify claim is readable
		got, err := store.Get("REQ-001")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if got.AgentID != "agent-1" {
			t.Errorf("Get AgentID = %q, want agent-1", got.AgentID)
		}

		// Release
		err = store.Release("REQ-001", "agent-1")
		if err != nil {
			t.Fatalf("Release failed: %v", err)
		}

		// Should be unclaimed now
		got, err = store.Get("REQ-001")
		if err != nil {
			t.Fatalf("Get after release failed: %v", err)
		}
		if got != nil {
			t.Error("expected nil claim after release")
		}
	})

	t.Run("double_claim_fails", func(t *testing.T) {
		store := newTestStore(t)

		_, err := store.Claim("REQ-001", "agent-1")
		if err != nil {
			t.Fatalf("first claim failed: %v", err)
		}

		_, err = store.Claim("REQ-001", "agent-2")
		if err == nil {
			t.Fatal("second claim should fail")
		}

		var ace *AlreadyClaimedError
		if !errors.As(err, &ace) {
			t.Fatalf("expected AlreadyClaimedError, got %T: %v", err, err)
		}
		if ace.HeldBy != "agent-1" {
			t.Errorf("HeldBy = %q, want agent-1", ace.HeldBy)
		}
	})

	t.Run("same_agent_double_claim_fails", func(t *testing.T) {
		store := newTestStore(t)

		_, err := store.Claim("REQ-001", "agent-1")
		if err != nil {
			t.Fatalf("first claim failed: %v", err)
		}

		_, err = store.Claim("REQ-001", "agent-1")
		if err == nil {
			t.Fatal("same-agent double claim should fail")
		}
	})

	t.Run("release_wrong_owner_fails", func(t *testing.T) {
		store := newTestStore(t)

		_, _ = store.Claim("REQ-001", "agent-1")

		err := store.Release("REQ-001", "agent-2")
		if err == nil {
			t.Fatal("release by non-owner should fail")
		}

		var noe *NotOwnerError
		if !errors.As(err, &noe) {
			t.Fatalf("expected NotOwnerError, got %T: %v", err, err)
		}
		if noe.Owner != "agent-1" || noe.Caller != "agent-2" {
			t.Errorf("NotOwnerError: owner=%q caller=%q", noe.Owner, noe.Caller)
		}
	})

	t.Run("release_unclaimed_fails", func(t *testing.T) {
		store := newTestStore(t)

		err := store.Release("REQ-NOPE", "agent-1")
		if err == nil {
			t.Fatal("release of unclaimed should fail")
		}

		var nce *NotClaimedError
		if !errors.As(err, &nce) {
			t.Fatalf("expected NotClaimedError, got %T: %v", err, err)
		}
	})

	t.Run("force_release", func(t *testing.T) {
		store := newTestStore(t)

		_, _ = store.Claim("REQ-001", "agent-1")

		err := store.ForceRelease("REQ-001")
		if err != nil {
			t.Fatalf("ForceRelease failed: %v", err)
		}

		got, _ := store.Get("REQ-001")
		if got != nil {
			t.Error("expected nil after force release")
		}
	})

	t.Run("list_claims", func(t *testing.T) {
		store := newTestStore(t)

		_, _ = store.Claim("REQ-001", "agent-1")
		_, _ = store.Claim("REQ-002", "agent-2")
		_, _ = store.Claim("REQ-003", "agent-1")

		claims, err := store.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(claims) != 3 {
			t.Errorf("expected 3 claims, got %d", len(claims))
		}
	})

	t.Run("get_unclaimed_returns_nil", func(t *testing.T) {
		store := newTestStore(t)

		got, err := store.Get("REQ-NOPE")
		if err != nil {
			t.Fatalf("Get unclaimed failed: %v", err)
		}
		if got != nil {
			t.Error("expected nil for unclaimed requirement")
		}
	})

	t.Run("concurrent_claims_no_corruption", func(t *testing.T) {
		store := newTestStore(t)

		// 10 agents race to claim the same requirement
		const agents = 10
		var wg sync.WaitGroup
		wins := make(chan string, agents)
		errs := make(chan error, agents)

		for i := 0; i < agents; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				agentID := "agent-" + string(rune('A'+id))
				_, err := store.Claim("REQ-RACE", agentID)
				if err == nil {
					wins <- agentID
				} else {
					errs <- err
				}
			}(i)
		}

		wg.Wait()
		close(wins)
		close(errs)

		// Exactly one agent should win
		winners := 0
		for range wins {
			winners++
		}
		if winners != 1 {
			t.Errorf("expected exactly 1 winner, got %d", winners)
		}

		// All errors should be AlreadyClaimedError
		for err := range errs {
			var ace *AlreadyClaimedError
			if !errors.As(err, &ace) {
				t.Errorf("expected AlreadyClaimedError, got %T: %v", err, err)
			}
		}

		// Claim file should be valid
		got, err := store.Get("REQ-RACE")
		if err != nil {
			t.Fatalf("Get after race failed: %v", err)
		}
		if got == nil {
			t.Fatal("expected claim to exist after race")
		}
	})

	t.Run("claim_persists_on_disk", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "claims")
		store, _ := NewClaimStore(dir)

		_, _ = store.Claim("REQ-001", "agent-1")

		// Create new store pointing at same directory
		store2, _ := NewClaimStore(dir)
		got, err := store2.Get("REQ-001")
		if err != nil {
			t.Fatalf("Get from second store failed: %v", err)
		}
		if got == nil || got.AgentID != "agent-1" {
			t.Error("claim should persist across store instances")
		}
	})
}

func newTestStore(t *testing.T) *ClaimStore {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "claims")
	store, err := NewClaimStore(dir)
	if err != nil {
		t.Fatalf("NewClaimStore failed: %v", err)
	}

	// Ensure the directory is clean (no leftover .DS_Store etc.)
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		_ = os.Remove(filepath.Join(dir, e.Name()))
	}

	return store
}

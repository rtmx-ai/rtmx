package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/results"
	"github.com/rtmx-ai/rtmx/internal/sync"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// ---------------------------------------------------------------------------
// REQ-SEC-001: Results file tampering -- no HMAC or signature validation
// ---------------------------------------------------------------------------
//
// VULNERABILITY: An attacker can craft a results.json file claiming all tests
// passed WITHOUT actually running any tests. The verify --results path trusts
// the file contents unconditionally because there is no HMAC, digital
// signature, or provenance check.
//
// This test PASSES today because the attack succeeds.
//
// FIXED behavior: verify should reject results that lack a valid HMAC
// computed over (results payload + database snapshot hash) using a
// project-level secret. Any unsigned or tampered file should be refused
// with a non-zero exit code and the database should remain unchanged.
// ---------------------------------------------------------------------------

func TestResultsTampering(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-001")

	// 1. Create a temp project directory with a minimal CSV database
	//    containing two MISSING requirements.
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "database.csv")

	db := database.NewDatabase()
	reqA := database.NewRequirement("REQ-AUTH-001")
	reqA.Category = "AUTH"
	reqA.RequirementText = "User login"
	reqA.Status = database.StatusMissing
	reqA.TestModule = "test_auth"
	reqA.TestFunction = "TestLogin"
	if err := db.Add(reqA); err != nil {
		t.Fatalf("failed to add REQ-AUTH-001: %v", err)
	}

	reqB := database.NewRequirement("REQ-AUTH-002")
	reqB.Category = "AUTH"
	reqB.RequirementText = "Password reset"
	reqB.Status = database.StatusMissing
	reqB.TestModule = "test_auth"
	reqB.TestFunction = "TestPasswordReset"
	if err := db.Add(reqB); err != nil {
		t.Fatalf("failed to add REQ-AUTH-002: %v", err)
	}

	if err := db.Save(dbPath); err != nil {
		t.Fatalf("failed to save database: %v", err)
	}

	// 2. Craft a fraudulent results.json -- tests "pass" without running.
	//    An attacker only needs to know requirement IDs (public in the CSV).
	crafted := []results.Result{
		{
			Marker: results.Marker{
				ReqID:    "REQ-AUTH-001",
				TestName: "TestLogin",
				TestFile: "fake_test.go",
			},
			Passed:   true,
			Duration: 0.001, // suspiciously fast
		},
		{
			Marker: results.Marker{
				ReqID:    "REQ-AUTH-002",
				TestName: "TestPasswordReset",
				TestFile: "fake_test.go",
			},
			Passed:   true,
			Duration: 0.001,
		},
	}

	resultsPath := filepath.Join(tmpDir, "results.json")
	data, err := json.MarshalIndent(crafted, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal crafted results: %v", err)
	}
	if err := os.WriteFile(resultsPath, data, 0644); err != nil {
		t.Fatalf("failed to write crafted results: %v", err)
	}

	// 3. Parse and apply the crafted results, replicating the verify logic.
	//    (We exercise the same code path as runVerifyFromResults.)
	f, err := os.Open(resultsPath)
	if err != nil {
		t.Fatalf("failed to open crafted results: %v", err)
	}
	defer func() { _ = f.Close() }()

	parsed, err := results.Parse(f)
	if err != nil {
		t.Fatalf("results.Parse should accept the crafted file: %v", err)
	}

	// No validation step rejects the file -- prove it.
	validationErrs := results.Validate(parsed)
	if len(validationErrs) > 0 {
		t.Fatalf("expected crafted results to pass validation, got %d errors", len(validationErrs))
	}

	// Group by requirement and compute status transitions.
	grouped := results.GroupByRequirement(parsed)
	for reqID, reqResults := range grouped {
		req := db.Get(reqID)
		if req == nil {
			t.Errorf("requirement %s not found in database", reqID)
			continue
		}

		allPassed := true
		for _, rr := range reqResults {
			if !rr.Passed {
				allPassed = false
			}
		}

		// Apply the same status transition logic as verify.go:
		// all passed -> COMPLETE
		if allPassed {
			req.Status = database.StatusComplete
		}
	}

	// 4. Assert: both requirements flipped to COMPLETE from crafted results.
	//    This PROVES the attack works -- no HMAC / signature check stopped it.
	for _, id := range []string{"REQ-AUTH-001", "REQ-AUTH-002"} {
		req := db.Get(id)
		if req == nil {
			t.Fatalf("%s disappeared from database", id)
		}
		if req.Status != database.StatusComplete {
			t.Errorf("ATTACK FAILED (unexpected): %s status is %s, expected COMPLETE",
				id, req.Status)
		}
	}

	// Confirm the crafted file had no HMAC field -- there is nothing to check.
	// When fixed, the results schema should require an "hmac" or "signature"
	// field and verify() should reject files without one.
	var raw []map[string]interface{}
	data2, _ := os.ReadFile(resultsPath)
	if err := json.Unmarshal(data2, &raw); err != nil {
		t.Fatalf("failed to re-parse results: %v", err)
	}
	for i, entry := range raw {
		if _, hasHMAC := entry["hmac"]; hasHMAC {
			t.Errorf("result[%d] unexpectedly has hmac field -- attack may no longer work", i)
		}
		if _, hasSig := entry["signature"]; hasSig {
			t.Errorf("result[%d] unexpectedly has signature field -- attack may no longer work", i)
		}
	}
}

// ---------------------------------------------------------------------------
// REQ-SEC-002: Sync protocol attacks -- no authentication, replay, injection
// ---------------------------------------------------------------------------
//
// VULNERABILITY: ApplyUpdates() accepts any RequirementUpdate with no
// authentication, no nonce/replay protection, and no authorization checks.
// An attacker who can deliver a SyncMessage can:
//   (a) Flip any requirement to COMPLETE without evidence.
//   (b) Replay the same message to re-apply changes (no idempotency guard).
//   (c) Inject phantom requirements that never existed.
//
// This test PASSES today because all three attacks succeed.
//
// FIXED behavior: Every SyncMessage should carry an HMAC (or JWT) signed
// by the sender. ApplyUpdates should verify the signature, reject replayed
// nonces (using a monotonic clock or nonce set), and reject requirement IDs
// not present in a pre-agreed manifest.
// ---------------------------------------------------------------------------

func TestSyncProtocolAttacks(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-002")

	// --- Setup: a database with one MISSING requirement ---
	db := database.NewDatabase()
	req := database.NewRequirement("REQ-CLI-001")
	req.Category = "CLI"
	req.RequirementText = "CLI help command"
	req.Status = database.StatusMissing
	if err := db.Add(req); err != nil {
		t.Fatalf("failed to add requirement: %v", err)
	}

	// ---- Attack (a): Unauthenticated status flip ----
	t.Run("unauthenticated_status_flip", func(t *testing.T) {
		// Craft an update that flips REQ-CLI-001 to COMPLETE.
		// No credentials, no token, no signature -- just raw data.
		updates := []sync.RequirementUpdate{
			{
				ReqID:     "REQ-CLI-001",
				Action:    "updated",
				Fields:    map[string]string{"status": "COMPLETE"},
				Source:    "attacker",
				Timestamp: time.Now(),
			},
		}

		result := sync.ApplyUpdates(db, updates)

		// The update should have been applied (vulnerability).
		if len(result.Updated) == 0 {
			t.Fatal("ATTACK FAILED (unexpected): update was rejected -- auth may exist")
		}

		got := db.Get("REQ-CLI-001")
		if got == nil {
			t.Fatal("REQ-CLI-001 vanished from database")
		}
		if got.Status != database.StatusComplete {
			t.Errorf("ATTACK FAILED (unexpected): status is %s, expected COMPLETE", got.Status)
		}

		// When fixed, ApplyUpdates should return an error or the result
		// should have entries in result.Errors indicating auth failure.
	})

	// Reset the requirement for the next sub-test.
	db.Get("REQ-CLI-001").Status = database.StatusMissing

	// ---- Attack (b): Replay attack ----
	t.Run("replay_attack", func(t *testing.T) {
		// Send the same message twice. If nonce checking existed, the
		// second application would be rejected.
		msg := sync.SyncMessage{
			Type: sync.MessageTypeUpdate,
			Room: "test-room",
			Updates: []sync.RequirementUpdate{
				{
					ReqID:     "REQ-CLI-001",
					Action:    "updated",
					Fields:    map[string]string{"status": "COMPLETE"},
					Source:    "attacker",
					Timestamp: time.Now(),
				},
			},
			Timestamp: time.Now(),
		}

		// First application.
		result1 := sync.ApplyUpdates(db, msg.Updates)
		if len(result1.Updated) == 0 {
			t.Fatal("first apply should succeed")
		}

		// Reset status so we can detect the replay succeeding.
		db.Get("REQ-CLI-001").Status = database.StatusMissing

		// Second application -- identical message, no nonce change.
		result2 := sync.ApplyUpdates(db, msg.Updates)
		if len(result2.Updated) == 0 {
			t.Fatal("ATTACK FAILED (unexpected): replay was rejected -- nonce checking may exist")
		}

		// Prove the replay changed state again.
		got := db.Get("REQ-CLI-001")
		if got.Status != database.StatusComplete {
			t.Errorf("ATTACK FAILED (unexpected): replay did not re-apply; status is %s", got.Status)
		}

		// When fixed, the second ApplyUpdates call should refuse the
		// message because its nonce/timestamp was already consumed.
	})

	// Reset for injection test.
	db.Get("REQ-CLI-001").Status = database.StatusMissing

	// ---- Attack (c): Phantom requirement injection ----
	t.Run("phantom_requirement_injection", func(t *testing.T) {
		// Inject a requirement that was never part of the project.
		// A real system should reject IDs not in a signed manifest.
		updates := []sync.RequirementUpdate{
			{
				ReqID:  "REQ-PHANTOM-999",
				Action: "added",
				Fields: map[string]string{
					"category":         "PHANTOM",
					"requirement_text": "I should not exist",
					"status":           "COMPLETE",
				},
				Source:    "attacker",
				Timestamp: time.Now(),
			},
		}

		result := sync.ApplyUpdates(db, updates)

		if len(result.Added) == 0 {
			t.Fatal("ATTACK FAILED (unexpected): phantom requirement was rejected")
		}

		phantom := db.Get("REQ-PHANTOM-999")
		if phantom == nil {
			t.Fatal("ATTACK FAILED (unexpected): phantom requirement not in database")
		}
		if phantom.Status != database.StatusComplete {
			t.Errorf("phantom requirement has status %s, expected COMPLETE", phantom.Status)
		}
		if phantom.Category != "PHANTOM" {
			t.Errorf("phantom requirement has category %q, expected PHANTOM", phantom.Category)
		}

		// When fixed, ApplyUpdates should refuse to add requirement IDs
		// that are not in the project's requirement manifest / whitelist.
	})
}

// ---------------------------------------------------------------------------
// REQ-SEC-003: Grant enforcement bypass -- grants are decorative
// ---------------------------------------------------------------------------
//
// VULNERABILITY: ConstraintAllows() correctly evaluates grant constraints,
// but ApplyUpdates() NEVER calls it. The grant system is purely cosmetic:
// constraints are defined in config but never enforced on the write path.
// An attacker with sync access can modify any requirement regardless of
// category/ID restrictions in their grant.
//
// This test PASSES today because the bypass succeeds.
//
// FIXED behavior: ApplyUpdates (or a wrapper) should accept a grant
// parameter and call ConstraintAllows for every update, rejecting any
// update that falls outside the grantee's permitted scope. The result
// should include entries in Errors for each rejected update.
// ---------------------------------------------------------------------------

func TestGrantEnforcementBypass(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-003")

	// --- Setup ---
	// Database has two requirements in different categories.
	db := database.NewDatabase()

	authReq := database.NewRequirement("REQ-AUTH-010")
	authReq.Category = "AUTH"
	authReq.RequirementText = "Token refresh"
	authReq.Status = database.StatusMissing
	if err := db.Add(authReq); err != nil {
		t.Fatalf("failed to add REQ-AUTH-010: %v", err)
	}

	cliReq := database.NewRequirement("REQ-CLI-010")
	cliReq.Category = "CLI"
	cliReq.RequirementText = "Help output formatting"
	cliReq.Status = database.StatusMissing
	if err := db.Add(cliReq); err != nil {
		t.Fatalf("failed to add REQ-CLI-010: %v", err)
	}

	// Grant: the collaborator "partner-a" may ONLY touch AUTH requirements.
	grant := config.SyncGrant{
		ID:      "grant-partner-a-001",
		Grantee: "partner-a",
		Role:    sync.RoleRequirementEditor,
		Constraints: config.GrantConstraint{
			Categories: []string{"AUTH"}, // Only AUTH is allowed.
		},
		CreatedAt: time.Now().Format(time.RFC3339),
		CreatedBy: "admin",
	}

	// Sanity check: ConstraintAllows correctly rejects CLI category.
	if sync.ConstraintAllows(grant.Constraints, cliReq) {
		t.Fatal("ConstraintAllows should reject CLI category for this grant")
	}
	// And correctly allows AUTH category.
	if !sync.ConstraintAllows(grant.Constraints, authReq) {
		t.Fatal("ConstraintAllows should allow AUTH category for this grant")
	}

	// --- Attack: modify a CLI requirement despite the AUTH-only grant ---
	t.Run("bypass_category_constraint", func(t *testing.T) {
		// The attacker (partner-a) crafts an update targeting REQ-CLI-010
		// which is OUTSIDE their granted AUTH category.
		updates := []sync.RequirementUpdate{
			{
				ReqID:  "REQ-CLI-010",
				Action: "updated",
				Fields: map[string]string{
					"status": "COMPLETE",
					"notes":  "Tampered by attacker outside grant scope",
				},
				Source:    "partner-a",
				Timestamp: time.Now(),
			},
		}

		// ApplyUpdates does not accept or check a grant -- it just applies.
		result := sync.ApplyUpdates(db, updates)

		// Prove the update was applied despite violating the grant.
		if len(result.Updated) == 0 {
			t.Fatal("ATTACK FAILED (unexpected): update to CLI req was rejected -- grant enforcement may exist")
		}
		if len(result.Errors) > 0 {
			t.Fatalf("ATTACK FAILED (unexpected): ApplyUpdates returned errors: %v", result.Errors)
		}

		got := db.Get("REQ-CLI-010")
		if got == nil {
			t.Fatal("REQ-CLI-010 disappeared")
		}
		if got.Status != database.StatusComplete {
			t.Errorf("ATTACK FAILED (unexpected): status is %s, want COMPLETE", got.Status)
		}
		if got.Notes != "Tampered by attacker outside grant scope" {
			t.Errorf("notes field was not tampered: %q", got.Notes)
		}

		// When fixed, ApplyUpdates should accept a *config.SyncGrant and
		// call ConstraintAllows for each update. The update to REQ-CLI-010
		// should be rejected with an error, and the requirement should
		// remain MISSING.
	})

	// --- Attack: modify a requirement via "added" action to bypass update checks ---
	t.Run("bypass_via_add_action", func(t *testing.T) {
		// Even if "updated" were checked, an attacker could try "added"
		// with an existing req_id. ApplyUpdates treats added+existing as
		// an update (see protocol.go lines 121-123).
		updates := []sync.RequirementUpdate{
			{
				ReqID:  "REQ-CLI-010",
				Action: "added", // Not "updated" -- tries to sidestep any future check
				Fields: map[string]string{
					"status": "COMPLETE",
					"notes":  "Bypassed via added action on existing requirement",
				},
				Source:    "partner-a",
				Timestamp: time.Now(),
			},
		}

		result := sync.ApplyUpdates(db, updates)

		// The "added" action on an existing requirement falls through
		// to applyFields -- no grant check occurs on either path.
		if len(result.Updated) == 0 && len(result.Added) == 0 {
			t.Fatal("ATTACK FAILED (unexpected): add-as-update was rejected")
		}

		got := db.Get("REQ-CLI-010")
		if got.Notes != "Bypassed via added action on existing requirement" {
			t.Errorf("expected notes to be tampered via add action, got: %q", got.Notes)
		}

		// When fixed, both "added" and "updated" actions should be
		// subject to grant constraint enforcement.
	})
}

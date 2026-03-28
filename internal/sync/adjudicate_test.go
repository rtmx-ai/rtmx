package sync

import (
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestAdjudicateRequirement_CategoryMatch_AutoApprove(t *testing.T) {
	rtmx.Req(t, "REQ-GO-077")

	rule := AdjudicationRule{
		Name:     "approve-sync",
		Category: "SYNC",
		Action:   ActionAutoApprove,
	}
	req := database.NewRequirement("REQ-SYNC-001")
	req.Category = "SYNC"

	decision := AdjudicateRequirement(rule, req)

	if decision.Action != ActionAutoApprove {
		t.Errorf("expected action %q, got %q", ActionAutoApprove, decision.Action)
	}
	if !decision.Matched {
		t.Error("expected decision to be matched")
	}
	if decision.RuleName != "approve-sync" {
		t.Errorf("expected rule name %q, got %q", "approve-sync", decision.RuleName)
	}
}

func TestAdjudicateRequirement_PriorityThreshold_RequireReview(t *testing.T) {
	rtmx.Req(t, "REQ-GO-077")

	rule := AdjudicationRule{
		Name:              "review-high-priority",
		PriorityThreshold: database.PriorityHigh,
		Action:            ActionRequireReview,
	}
	req := database.NewRequirement("REQ-TEST-001")
	req.Priority = database.PriorityP0 // P0 >= HIGH threshold

	decision := AdjudicateRequirement(rule, req)

	if decision.Action != ActionRequireReview {
		t.Errorf("expected action %q, got %q", ActionRequireReview, decision.Action)
	}
	if !decision.Matched {
		t.Error("expected decision to be matched")
	}
}

func TestAdjudicateRequirement_PriorityBelowThreshold_NoMatch(t *testing.T) {
	rtmx.Req(t, "REQ-GO-077")

	rule := AdjudicationRule{
		Name:              "review-high-priority",
		PriorityThreshold: database.PriorityHigh,
		Action:            ActionRequireReview,
	}
	req := database.NewRequirement("REQ-TEST-001")
	req.Priority = database.PriorityLow // LOW < HIGH threshold

	decision := AdjudicateRequirement(rule, req)

	if decision.Matched {
		t.Error("expected decision to NOT match for low priority")
	}
}

func TestAdjudicateRequirement_PhaseMatch_Reject(t *testing.T) {
	rtmx.Req(t, "REQ-GO-077")

	rule := AdjudicationRule{
		Name:   "reject-phase-99",
		Phase:  99,
		Action: ActionReject,
	}
	req := database.NewRequirement("REQ-TEST-001")
	req.Phase = 99

	decision := AdjudicateRequirement(rule, req)

	if decision.Action != ActionReject {
		t.Errorf("expected action %q, got %q", ActionReject, decision.Action)
	}
	if !decision.Matched {
		t.Error("expected decision to be matched")
	}
}

func TestAdjudicateRequirement_CategoryMismatch_NoMatch(t *testing.T) {
	rtmx.Req(t, "REQ-GO-077")

	rule := AdjudicationRule{
		Name:     "approve-sync",
		Category: "SYNC",
		Action:   ActionAutoApprove,
	}
	req := database.NewRequirement("REQ-AUTH-001")
	req.Category = "AUTH"

	decision := AdjudicateRequirement(rule, req)

	if decision.Matched {
		t.Error("expected decision to NOT match for wrong category")
	}
}

func TestEvaluateAll_NoMatchingRule_DefaultRequireReview(t *testing.T) {
	rtmx.Req(t, "REQ-GO-077")

	rules := []AdjudicationRule{
		{
			Name:     "approve-sync",
			Category: "SYNC",
			Action:   ActionAutoApprove,
		},
	}
	req := database.NewRequirement("REQ-AUTH-001")
	req.Category = "AUTH"

	decisions := EvaluateAll(rules, []*database.Requirement{req})

	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}
	if decisions[0].Action != ActionRequireReview {
		t.Errorf("expected default action %q, got %q", ActionRequireReview, decisions[0].Action)
	}
	if decisions[0].Reason != "no matching rule; defaulting to require-review" {
		t.Errorf("unexpected reason: %q", decisions[0].Reason)
	}
}

func TestEvaluateAll_FirstMatchWins(t *testing.T) {
	rtmx.Req(t, "REQ-GO-077")

	rules := []AdjudicationRule{
		{
			Name:     "reject-sync",
			Category: "SYNC",
			Action:   ActionReject,
		},
		{
			Name:     "approve-sync",
			Category: "SYNC",
			Action:   ActionAutoApprove,
		},
	}
	req := database.NewRequirement("REQ-SYNC-001")
	req.Category = "SYNC"

	decisions := EvaluateAll(rules, []*database.Requirement{req})

	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}
	// First matching rule (reject-sync) should win over approve-sync
	if decisions[0].Action != ActionReject {
		t.Errorf("expected first-match action %q, got %q", ActionReject, decisions[0].Action)
	}
	if decisions[0].RuleName != "reject-sync" {
		t.Errorf("expected rule name %q, got %q", "reject-sync", decisions[0].RuleName)
	}
}

func TestEvaluateAll_MultipleRequirements(t *testing.T) {
	rtmx.Req(t, "REQ-GO-077")

	rules := []AdjudicationRule{
		{
			Name:     "approve-sync",
			Category: "SYNC",
			Action:   ActionAutoApprove,
		},
		{
			Name:              "review-high",
			PriorityThreshold: database.PriorityHigh,
			Action:            ActionRequireReview,
		},
	}
	syncReq := database.NewRequirement("REQ-SYNC-001")
	syncReq.Category = "SYNC"
	syncReq.Priority = database.PriorityLow

	highReq := database.NewRequirement("REQ-AUTH-001")
	highReq.Category = "AUTH"
	highReq.Priority = database.PriorityP0

	decisions := EvaluateAll(rules, []*database.Requirement{syncReq, highReq})

	if len(decisions) != 2 {
		t.Fatalf("expected 2 decisions, got %d", len(decisions))
	}
	if decisions[0].Action != ActionAutoApprove {
		t.Errorf("syncReq: expected %q, got %q", ActionAutoApprove, decisions[0].Action)
	}
	if decisions[1].Action != ActionRequireReview {
		t.Errorf("highReq: expected %q, got %q", ActionRequireReview, decisions[1].Action)
	}
}

func TestLoadRules_FromConfig(t *testing.T) {
	rtmx.Req(t, "REQ-GO-077")

	cfg := &config.AdjudicationConfig{
		Enabled: true,
		Rules: []config.AdjudicationRuleConfig{
			{
				Name:     "approve-sync",
				Category: "SYNC",
				Action:   "auto-approve",
			},
			{
				Name:              "review-high",
				PriorityThreshold: "HIGH",
				Action:            "require-review",
			},
			{
				Name:   "reject-phase-99",
				Phase:  99,
				Action: "reject",
			},
		},
	}

	rules, err := LoadRules(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}

	if rules[0].Category != "SYNC" {
		t.Errorf("rule 0: expected category SYNC, got %q", rules[0].Category)
	}
	if rules[0].Action != ActionAutoApprove {
		t.Errorf("rule 0: expected action %q, got %q", ActionAutoApprove, rules[0].Action)
	}

	if rules[1].PriorityThreshold != database.PriorityHigh {
		t.Errorf("rule 1: expected priority HIGH, got %q", rules[1].PriorityThreshold)
	}
	if rules[1].Action != ActionRequireReview {
		t.Errorf("rule 1: expected action %q, got %q", ActionRequireReview, rules[1].Action)
	}

	if rules[2].Phase != 99 {
		t.Errorf("rule 2: expected phase 99, got %d", rules[2].Phase)
	}
	if rules[2].Action != ActionReject {
		t.Errorf("rule 2: expected action %q, got %q", ActionReject, rules[2].Action)
	}
}

func TestLoadRules_InvalidAction(t *testing.T) {
	rtmx.Req(t, "REQ-GO-077")

	cfg := &config.AdjudicationConfig{
		Enabled: true,
		Rules: []config.AdjudicationRuleConfig{
			{
				Name:   "bad-rule",
				Action: "invalid-action",
			},
		},
	}

	_, err := LoadRules(cfg)
	if err == nil {
		t.Fatal("expected error for invalid action, got nil")
	}
}

func TestLoadRules_InvalidPriority(t *testing.T) {
	rtmx.Req(t, "REQ-GO-077")

	cfg := &config.AdjudicationConfig{
		Enabled: true,
		Rules: []config.AdjudicationRuleConfig{
			{
				Name:              "bad-priority",
				PriorityThreshold: "ULTRA",
				Action:            "reject",
			},
		},
	}

	_, err := LoadRules(cfg)
	if err == nil {
		t.Fatal("expected error for invalid priority, got nil")
	}
}

func TestAdjudicateRequirement_CombinedConditions(t *testing.T) {
	rtmx.Req(t, "REQ-GO-077")

	rule := AdjudicationRule{
		Name:              "approve-sync-low",
		Category:          "SYNC",
		PriorityThreshold: database.PriorityMedium,
		Phase:             3,
		Action:            ActionAutoApprove,
	}

	// All conditions match
	req := database.NewRequirement("REQ-SYNC-001")
	req.Category = "SYNC"
	req.Priority = database.PriorityMedium
	req.Phase = 3

	decision := AdjudicateRequirement(rule, req)
	if !decision.Matched {
		t.Error("expected match when all conditions satisfied")
	}

	// Category mismatch
	req2 := database.NewRequirement("REQ-AUTH-001")
	req2.Category = "AUTH"
	req2.Priority = database.PriorityMedium
	req2.Phase = 3

	decision2 := AdjudicateRequirement(rule, req2)
	if decision2.Matched {
		t.Error("expected no match when category mismatches")
	}
}

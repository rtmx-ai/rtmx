package sync

import (
	"fmt"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
)

// Action represents the adjudication outcome for a requirement.
type Action string

const (
	// ActionAutoApprove means the requirement is automatically approved.
	ActionAutoApprove Action = "auto-approve"

	// ActionRequireReview means the requirement needs human review.
	ActionRequireReview Action = "require-review"

	// ActionReject means the requirement is rejected.
	ActionReject Action = "reject"
)

// parseAction converts a string to a valid Action.
func parseAction(s string) (Action, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "auto-approve":
		return ActionAutoApprove, nil
	case "require-review":
		return ActionRequireReview, nil
	case "reject":
		return ActionReject, nil
	default:
		return "", fmt.Errorf("invalid adjudication action: %q (valid: auto-approve, require-review, reject)", s)
	}
}

// AdjudicationRule defines conditions and an action for evaluating
// incoming requirement PRs.
type AdjudicationRule struct {
	// Name identifies the rule for reporting purposes.
	Name string

	// Category matches requirements by category (exact match, case-sensitive).
	// Empty means no category filter.
	Category string

	// PriorityThreshold matches requirements with priority at or above
	// this threshold (lower Weight = higher priority).
	// Zero value (empty) means no priority filter.
	PriorityThreshold database.Priority

	// Phase matches requirements in a specific phase.
	// Zero means no phase filter.
	Phase int

	// Action is the outcome when all conditions match.
	Action Action
}

// Decision is the result of evaluating a requirement against an adjudication rule.
type Decision struct {
	// ReqID is the requirement that was evaluated.
	ReqID string

	// Action is the adjudication outcome.
	Action Action

	// Reason explains why this decision was made.
	Reason string

	// RuleName identifies which rule produced this decision.
	RuleName string

	// Matched indicates whether a rule matched.
	Matched bool
}

// AdjudicateRequirement evaluates a single requirement against a single rule.
// If all non-empty conditions on the rule match the requirement, the decision
// is marked as Matched with the rule's action. Otherwise Matched is false.
func AdjudicateRequirement(rule AdjudicationRule, req *database.Requirement) Decision {
	decision := Decision{
		ReqID:    req.ReqID,
		RuleName: rule.Name,
	}

	// Check category condition
	if rule.Category != "" && req.Category != rule.Category {
		return decision
	}

	// Check priority threshold condition: requirement priority must be
	// at or above the threshold (lower weight = higher priority).
	if rule.PriorityThreshold != "" {
		if req.Priority.Weight() > rule.PriorityThreshold.Weight() {
			return decision
		}
	}

	// Check phase condition
	if rule.Phase != 0 && req.Phase != rule.Phase {
		return decision
	}

	decision.Matched = true
	decision.Action = rule.Action
	decision.Reason = fmt.Sprintf("matched rule %q", rule.Name)
	return decision
}

// EvaluateAll evaluates each requirement against the ordered list of rules.
// For each requirement, the first matching rule wins. If no rule matches,
// the default action is require-review.
func EvaluateAll(rules []AdjudicationRule, requirements []*database.Requirement) []Decision {
	decisions := make([]Decision, 0, len(requirements))

	for _, req := range requirements {
		decided := false
		for _, rule := range rules {
			d := AdjudicateRequirement(rule, req)
			if d.Matched {
				decisions = append(decisions, d)
				decided = true
				break
			}
		}
		if !decided {
			decisions = append(decisions, Decision{
				ReqID:  req.ReqID,
				Action: ActionRequireReview,
				Reason: "no matching rule; defaulting to require-review",
			})
		}
	}

	return decisions
}

// LoadRules converts configuration-level adjudication rules into
// the engine's AdjudicationRule slice. It validates actions and
// priority thresholds, returning an error on invalid input.
func LoadRules(cfg *config.AdjudicationConfig) ([]AdjudicationRule, error) {
	rules := make([]AdjudicationRule, 0, len(cfg.Rules))

	for i, rc := range cfg.Rules {
		action, err := parseAction(rc.Action)
		if err != nil {
			return nil, fmt.Errorf("rule %d (%q): %w", i, rc.Name, err)
		}

		rule := AdjudicationRule{
			Name:     rc.Name,
			Category: rc.Category,
			Phase:    rc.Phase,
			Action:   action,
		}

		if rc.PriorityThreshold != "" {
			p, err := database.ParsePriority(rc.PriorityThreshold)
			if err != nil {
				return nil, fmt.Errorf("rule %d (%q): invalid priority_threshold: %w", i, rc.Name, err)
			}
			rule.PriorityThreshold = p
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

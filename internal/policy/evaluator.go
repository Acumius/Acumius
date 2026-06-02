package policy

import (
	"context"
	"fmt"
	"strings"
)

// Store defines the interface for retrieving policies from persistent storage.
type Store interface {
	GetPolicies(ctx context.Context, agentDID string) ([]Policy, error)
}

// Evaluator checks actions against policies.
type Evaluator struct {
	cache Cache
	store Store
}

// NewEvaluator creates a new policy Evaluator.
func NewEvaluator(cache Cache, store Store) *Evaluator {
	return &Evaluator{
		cache: cache,
		store: store,
	}
}

// Request represents an access control request.
type Request struct {
	AgentDID   string
	Action     string
	MemoryType string
	Namespace  string
}

// Result is the result of policy evaluation.
type Result struct {
	Allowed bool
	Reason  string
}

// Evaluate determines whether a request is allowed based on the agent's policies.
func (e *Evaluator) Evaluate(ctx context.Context, req Request) (Result, error) {
	policies, err := e.getPolicies(ctx, req.AgentDID)
	if err != nil {
		return Result{Allowed: false, Reason: "failed to retrieve policies"}, err
	}

	if len(policies) == 0 {
		return Result{Allowed: false, Reason: "no policies found (default deny)"}, nil
	}

	// Fail-closed default
	allowed := false
	reason := "no matching allow rule"

	for _, p := range policies {
		for _, rule := range p.Content.MemoryRules {
			if !actionMatches(rule.Action, req.Action) {
				continue
			}
			if !stringMatches(rule.TargetTypes, req.MemoryType) {
				continue
			}
			if !stringMatches(rule.Namespaces, req.Namespace) {
				continue
			}

			if rule.Effect == "deny" {
				return Result{Allowed: false, Reason: "explicit deny by policy " + p.ID}, nil
			}

			if rule.Effect == "allow" {
				allowed = true
				reason = "allowed by policy " + p.ID
			}
		}
	}

	return Result{Allowed: allowed, Reason: reason}, nil
}

// CheckPII determines what to do with a given PII type according to policy.
func (e *Evaluator) CheckPII(ctx context.Context, agentDID string, piiType string) (string, error) {
	policies, err := e.getPolicies(ctx, agentDID)
	if err != nil {
		return "block", err // safe default
	}

	action := "allow" // default

	for _, p := range policies {
		for _, rule := range p.Content.PIIRules {
			if stringMatches(rule.TypesToRedact, piiType) {
				// Priority: block > redact > allow
				if rule.Action == "block" {
					return "block", nil
				} else if rule.Action == "redact" {
					action = "redact"
				}
			}
		}
	}

	return action, nil
}

func (e *Evaluator) getPolicies(ctx context.Context, agentDID string) ([]Policy, error) {
	// Try cache first
	if e.cache != nil {
		cached, err := e.cache.GetPolicies(ctx, agentDID)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	// Fallback to store
	policies, err := e.store.GetPolicies(ctx, agentDID)
	if err != nil {
		return nil, fmt.Errorf("store get: %w", err)
	}

	// Update cache
	if e.cache != nil {
		_ = e.cache.SetPolicies(ctx, agentDID, policies)
	}

	return policies, nil
}

func actionMatches(ruleAction, reqAction string) bool {
	if ruleAction == "*" {
		return true
	}
	return strings.EqualFold(ruleAction, reqAction)
}

func stringMatches(list []string, item string) bool {
	for _, val := range list {
		if val == "*" || strings.EqualFold(val, item) {
			return true
		}
	}
	return false
}

package policy_test

import (
	"context"
	"testing"

	"github.com/Acumius/Acumius/internal/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	policies []policy.Policy
}

func (m *mockStore) GetPolicies(ctx context.Context, agentDID string) ([]policy.Policy, error) {
	var match []policy.Policy
	for _, p := range m.policies {
		if p.AgentDID == agentDID {
			match = append(match, p)
		}
	}
	return match, nil
}

func (m *mockStore) SavePolicy(ctx context.Context, p policy.Policy) error {
	m.policies = append(m.policies, p)
	return nil
}

func TestEvaluator_Evaluate_Allow(t *testing.T) {
	store := &mockStore{
		policies: []policy.Policy{
			{
				ID:       "pol-1",
				AgentDID: "did:test:1",
				Content: policy.Content{
					MemoryRules: []policy.MemoryRule{
						{
							Action:      "read",
							TargetTypes: []string{"chat"},
							Namespaces:  []string{"*"},
							Effect:      "allow",
						},
					},
				},
			},
		},
	}

	// No cache, use mock store
	eval := policy.NewEvaluator(nil, store)

	res, err := eval.Evaluate(context.Background(), policy.Request{
		AgentDID:   "did:test:1",
		Action:     "read",
		MemoryType: "chat",
		Namespace:  "default",
	})

	require.NoError(t, err)
	assert.True(t, res.Allowed)
	assert.Contains(t, res.Reason, "allowed by policy pol-1")
}

func TestEvaluator_Evaluate_DenyByDefault(t *testing.T) {
	store := &mockStore{
		policies: []policy.Policy{
			{
				ID:       "pol-2",
				AgentDID: "did:test:2",
				Content: policy.Content{
					MemoryRules: []policy.MemoryRule{
						{
							Action:      "write",
							TargetTypes: []string{"chat"},
							Namespaces:  []string{"*"},
							Effect:      "allow",
						},
					},
				},
			},
		},
	}

	eval := policy.NewEvaluator(nil, store)

	res, err := eval.Evaluate(context.Background(), policy.Request{
		AgentDID:   "did:test:2",
		Action:     "read", // policy only allows write
		MemoryType: "chat",
		Namespace:  "default",
	})

	require.NoError(t, err)
	assert.False(t, res.Allowed)
	assert.Equal(t, "no matching allow rule", res.Reason)
}

func TestEvaluator_Evaluate_ExplicitDeny(t *testing.T) {
	store := &mockStore{
		policies: []policy.Policy{
			{
				ID:       "pol-3",
				AgentDID: "did:test:3",
				Content: policy.Content{
					MemoryRules: []policy.MemoryRule{
						{
							Action:      "read",
							TargetTypes: []string{"finance"},
							Namespaces:  []string{"*"},
							Effect:      "deny",
						},
						{
							Action:      "read",
							TargetTypes: []string{"*"},
							Namespaces:  []string{"*"},
							Effect:      "allow",
						},
					},
				},
			},
		},
	}

	eval := policy.NewEvaluator(nil, store)

	res, err := eval.Evaluate(context.Background(), policy.Request{
		AgentDID:   "did:test:3",
		Action:     "read",
		MemoryType: "finance",
		Namespace:  "default",
	})

	require.NoError(t, err)
	assert.False(t, res.Allowed)
	assert.Contains(t, res.Reason, "explicit deny")
}

func TestEvaluator_CheckPII(t *testing.T) {
	store := &mockStore{
		policies: []policy.Policy{
			{
				ID:       "pol-4",
				AgentDID: "did:test:4",
				Content: policy.Content{
					PIIRules: []policy.PIIRule{
						{
							TypesToRedact: []string{"ssn"},
							Action:        "block",
						},
						{
							TypesToRedact: []string{"email"},
							Action:        "redact",
						},
					},
				},
			},
		},
	}

	eval := policy.NewEvaluator(nil, store)

	// Test block
	action, err := eval.CheckPII(context.Background(), "did:test:4", "ssn")
	require.NoError(t, err)
	assert.Equal(t, "block", action)

	// Test redact
	action, err = eval.CheckPII(context.Background(), "did:test:4", "email")
	require.NoError(t, err)
	assert.Equal(t, "redact", action)

	// Test allow (default)
	action, err = eval.CheckPII(context.Background(), "did:test:4", "name")
	require.NoError(t, err)
	assert.Equal(t, "allow", action)
}

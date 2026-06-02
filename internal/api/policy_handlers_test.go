package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Acumius/Acumius/internal/api"
	"github.com/Acumius/Acumius/internal/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPolicyStore struct {
	policies []policy.Policy
}

func (m *mockPolicyStore) GetPolicies(ctx context.Context, agentDID string) ([]policy.Policy, error) {
	var match []policy.Policy
	for _, p := range m.policies {
		if p.AgentDID == agentDID {
			match = append(match, p)
		}
	}
	return match, nil
}

func (m *mockPolicyStore) SavePolicy(ctx context.Context, p policy.Policy) error {
	m.policies = append(m.policies, p)
	return nil
}

func TestPolicyHandler_CreatePolicy(t *testing.T) {
	store := &mockPolicyStore{}
	eval := policy.NewEvaluator(nil, store)
	handler := api.NewPolicyHandler(store, eval)

	p := policy.Policy{
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
	}
	body, _ := json.Marshal(p)

	req := httptest.NewRequest(http.MethodPost, "/api/policies", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.CreatePolicy(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]string
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "created", resp["status"])
	assert.Equal(t, "pol-1", resp["id"])
	assert.Len(t, store.policies, 1)
}

func TestPolicyHandler_EvaluatePolicy(t *testing.T) {
	store := &mockPolicyStore{
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
	eval := policy.NewEvaluator(nil, store)
	handler := api.NewPolicyHandler(store, eval)

	preq := policy.Request{
		AgentDID:   "did:test:1",
		Action:     "read",
		MemoryType: "chat",
		Namespace:  "default",
	}
	body, _ := json.Marshal(preq)

	req := httptest.NewRequest(http.MethodPost, "/api/policies/evaluate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.EvaluatePolicy(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp policy.Result
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Allowed)
}

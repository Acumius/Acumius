package policy_test

import (
	"testing"

	"github.com/Acumius/Acumius/internal/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePolicy_JSON(t *testing.T) {
	jsonData := []byte(`{
		"id": "pol-123",
		"agent_did": "did:example:123",
		"content": {
			"memory_rules": [
				{
					"action": "read",
					"target_types": ["*"],
					"namespaces": ["*"],
					"effect": "allow"
				}
			]
		}
	}`)

	p, err := policy.ParsePolicy(jsonData)
	require.NoError(t, err)
	assert.Equal(t, "pol-123", p.ID)
	assert.Equal(t, "did:example:123", p.AgentDID)
	require.Len(t, p.Content.MemoryRules, 1)
	assert.Equal(t, "allow", p.Content.MemoryRules[0].Effect)
}

func TestParsePolicy_YAML(t *testing.T) {
	yamlData := []byte(`
id: pol-456
agent_did: did:example:456
content:
  memory_rules:
    - action: write
      target_types: ["document"]
      namespaces: ["public"]
      effect: deny
  pii_rules:
    - types_to_redact: ["email"]
      action: redact
`)

	p, err := policy.ParsePolicy(yamlData)
	require.NoError(t, err)
	assert.Equal(t, "pol-456", p.ID)
	require.Len(t, p.Content.MemoryRules, 1)
	assert.Equal(t, "deny", p.Content.MemoryRules[0].Effect)
	require.Len(t, p.Content.PIIRules, 1)
	assert.Equal(t, "redact", p.Content.PIIRules[0].Action)
}

func TestParsePolicy_InvalidContent(t *testing.T) {
	invalidJSON := []byte(`{
		"content": {
			"memory_rules": [
				{
					"action": "read",
					"target_types": ["*"],
					"namespaces": ["*"],
					"effect": "invalid_effect"
				}
			]
		}
	}`)

	_, err := policy.ParsePolicy(invalidJSON)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid policy content")
	assert.Contains(t, err.Error(), "must be 'allow' or 'deny'")
}

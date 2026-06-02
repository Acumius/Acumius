package policy

import (
	"encoding/json"
	"time"
)

// Policy represents the declarative rules an agent or organization sets for interactions.
type Policy struct {
	ID        string    `json:"id"`
	AgentDID  string    `json:"agent_did"`
	Version   string    `json:"version"`
	Content   Content   `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Content holds the actual rules of the policy.
type Content struct {
	MemoryRules     []MemoryRule     `json:"memory_rules,omitempty" yaml:"memory_rules,omitempty"`
	DelegationRules []DelegationRule `json:"delegation_rules,omitempty" yaml:"delegation_rules,omitempty"`
	PIIRules        []PIIRule        `json:"pii_rules,omitempty" yaml:"pii_rules,omitempty"`
	AuditRules      []AuditRule      `json:"audit_rules,omitempty" yaml:"audit_rules,omitempty"`
}

// MemoryRule dictates access to different memory types and namespaces.
type MemoryRule struct {
	Action      string   `json:"action" yaml:"action"`             // e.g., "read", "write", "search", "*"
	TargetTypes []string `json:"target_types" yaml:"target_types"` // e.g., ["working", "episodic"]
	Namespaces  []string `json:"namespaces" yaml:"namespaces"`     // e.g., ["project-x", "*"]
	Effect      string   `json:"effect" yaml:"effect"`             // "allow", "deny"
}

// DelegationRule dictates how deep delegations can go and the required reputation.
type DelegationRule struct {
	MaxDepth       int      `json:"max_depth" yaml:"max_depth"`
	MinReputation  int      `json:"min_reputation" yaml:"min_reputation"`
	AllowedAgents  []string `json:"allowed_agents,omitempty" yaml:"allowed_agents,omitempty"`
	ExcludedAgents []string `json:"excluded_agents,omitempty" yaml:"excluded_agents,omitempty"`
	Effect         string   `json:"effect" yaml:"effect"` // "allow", "deny"
}

// PIIRule dictates how PII is handled.
type PIIRule struct {
	TypesToRedact []string `json:"types_to_redact" yaml:"types_to_redact"` // e.g., ["email", "ssn"]
	Action        string   `json:"action" yaml:"action"`                   // "redact", "block", "allow"
}

// AuditRule dictates what actions should be heavily audited.
type AuditRule struct {
	ActionsToLog []string `json:"actions_to_log" yaml:"actions_to_log"` // e.g., ["read_working_memory", "*"]
	Level        string   `json:"level" yaml:"level"`                   // "basic", "detailed"
}

// ToJSON converts Policy Content to a JSONRawMessage for database storage.
func (c *Content) ToJSON() (json.RawMessage, error) {
	bytes, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(bytes), nil
}

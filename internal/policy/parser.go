package policy

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParsePolicy parses a policy string (YAML or JSON) into a Policy struct.
// Note: It primarily parses the Content section of a policy.
func ParsePolicy(data []byte) (*Policy, error) {
	var p Policy

	// Try JSON first
	if isJSON(data) {
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("failed to parse JSON policy: %w", err)
		}
	} else {
		// Fallback to YAML
		if err := yaml.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("failed to parse YAML policy: %w", err)
		}
	}

	if err := validateContent(&p.Content); err != nil {
		return nil, fmt.Errorf("invalid policy content: %w", err)
	}

	return &p, nil
}

// ParseContent parses just the Content part of the policy (YAML or JSON).
func ParseContent(data []byte) (*Content, error) {
	var c Content

	if isJSON(data) {
		if err := json.Unmarshal(data, &c); err != nil {
			return nil, fmt.Errorf("failed to parse JSON content: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(data, &c); err != nil {
			return nil, fmt.Errorf("failed to parse YAML content: %w", err)
		}
	}

	if err := validateContent(&c); err != nil {
		return nil, fmt.Errorf("invalid policy content: %w", err)
	}

	return &c, nil
}

// isJSON checks if the byte slice looks like a JSON object.
func isJSON(data []byte) bool {
	str := strings.TrimSpace(string(data))
	return strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}")
}

func validateContent(c *Content) error {
	for _, rule := range c.MemoryRules {
		if rule.Action == "" || len(rule.TargetTypes) == 0 || len(rule.Namespaces) == 0 || rule.Effect == "" {
			return fmt.Errorf("memory rule missing required fields")
		}
		if rule.Effect != "allow" && rule.Effect != "deny" {
			return fmt.Errorf("memory rule effect must be 'allow' or 'deny'")
		}
	}

	for _, rule := range c.DelegationRules {
		if rule.Effect != "allow" && rule.Effect != "deny" {
			return fmt.Errorf("delegation rule effect must be 'allow' or 'deny'")
		}
	}

	for _, rule := range c.PIIRules {
		if rule.Action != "redact" && rule.Action != "block" && rule.Action != "allow" {
			return fmt.Errorf("pii rule action must be 'redact', 'block', or 'allow'")
		}
	}

	for _, rule := range c.AuditRules {
		if rule.Level != "basic" && rule.Level != "detailed" {
			return fmt.Errorf("audit rule level must be 'basic' or 'detailed'")
		}
	}

	return nil
}

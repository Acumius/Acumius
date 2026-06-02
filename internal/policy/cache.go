package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/valkey-io/valkey-go"
)

// Cache interface defines the methods for retrieving and caching policies.
type Cache interface {
	GetPolicies(ctx context.Context, agentDID string) ([]Policy, error)
	SetPolicies(ctx context.Context, agentDID string, policies []Policy) error
	Invalidate(ctx context.Context, agentDID string) error
}

// ValkeyCache is a Valkey-backed implementation of Cache.
type ValkeyCache struct {
	client valkey.Client
	ttl    time.Duration
}

// NewValkeyCache creates a new ValkeyCache.
func NewValkeyCache(client valkey.Client) *ValkeyCache {
	return &ValkeyCache{
		client: client,
		ttl:    5 * time.Minute,
	}
}

func (c *ValkeyCache) key(agentDID string) string {
	return fmt.Sprintf("acumius:policy:cache:%s", agentDID)
}

// GetPolicies retrieves cached policies for an agent.
func (c *ValkeyCache) GetPolicies(ctx context.Context, agentDID string) ([]Policy, error) {
	data, err := c.client.Do(ctx, c.client.B().Get().Key(c.key(agentDID)).Build()).AsBytes()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("valkey get: %w", err)
	}

	var policies []Policy
	if err := json.Unmarshal(data, &policies); err != nil {
		return nil, fmt.Errorf("unmarshal policies: %w", err)
	}

	return policies, nil
}

// SetPolicies caches policies for an agent.
func (c *ValkeyCache) SetPolicies(ctx context.Context, agentDID string, policies []Policy) error {
	data, err := json.Marshal(policies)
	if err != nil {
		return fmt.Errorf("marshal policies: %w", err)
	}

	err = c.client.Do(ctx, c.client.B().Set().Key(c.key(agentDID)).Value(string(data)).ExSeconds(int64(c.ttl.Seconds())).Build()).Error()
	if err != nil {
		return fmt.Errorf("valkey set: %w", err)
	}

	return nil
}

// Invalidate removes an agent's policies from the cache.
func (c *ValkeyCache) Invalidate(ctx context.Context, agentDID string) error {
	err := c.client.Do(ctx, c.client.B().Del().Key(c.key(agentDID)).Build()).Error()
	if err != nil {
		return fmt.Errorf("valkey del: %w", err)
	}
	return nil
}

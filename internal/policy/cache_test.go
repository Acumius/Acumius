package policy_test

import (
	"context"
	"testing"

	"github.com/Acumius/Acumius/internal/policy"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valkey-io/valkey-go"
)

func TestValkeyCache(t *testing.T) {
	// Start miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Connect valkey client
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress:  []string{mr.Addr()},
		DisableCache: true,
	})
	require.NoError(t, err)
	defer client.Close()

	cache := policy.NewValkeyCache(client)
	ctx := context.Background()

	agentDID := "did:example:cache-123"

	// 1. Should return nil when no policies exist
	policies, err := cache.GetPolicies(ctx, agentDID)
	require.NoError(t, err)
	assert.Nil(t, policies)

	// 2. Set policies
	testPolicies := []policy.Policy{
		{
			ID:       "pol-cache-1",
			AgentDID: agentDID,
			Content: policy.Content{
				MemoryRules: []policy.MemoryRule{
					{
						Action: "read",
						Effect: "allow",
					},
				},
			},
		},
	}

	err = cache.SetPolicies(ctx, agentDID, testPolicies)
	require.NoError(t, err)

	// 3. Get policies
	cachedPolicies, err := cache.GetPolicies(ctx, agentDID)
	require.NoError(t, err)
	require.Len(t, cachedPolicies, 1)
	assert.Equal(t, "pol-cache-1", cachedPolicies[0].ID)
	assert.Equal(t, "allow", cachedPolicies[0].Content.MemoryRules[0].Effect)

	// 4. Invalidate
	err = cache.Invalidate(ctx, agentDID)
	require.NoError(t, err)

	// 5. Verify invalidated
	policies, err = cache.GetPolicies(ctx, agentDID)
	require.NoError(t, err)
	assert.Nil(t, policies)
}

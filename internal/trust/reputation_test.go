package trust

import (
	"context"
	"testing"
	"time"
)

func TestCalculateReputation(t *testing.T) {
	tests := []struct {
		name     string
		factors  ReputationFactors
		expected int
	}{
		{
			name:     "default state",
			factors:  ReputationFactors{CompletionRate: 1.0},
			expected: 700, // 500 + 200
		},
		{
			name: "high reputation",
			factors: ReputationFactors{
				CompletionRate:     1.0,
				PeerVerifications:  5,
				MemoryAttestations: 5,
			},
			expected: 1000, // 500 + 200 + 250 + 125 = 1075 -> clamped to 1000
		},
		{
			name: "low reputation with violations",
			factors: ReputationFactors{
				CompletionRate:     0.5,
				PeerVerifications:  1,
				MemoryAttestations: 2,
				PolicyViolations:   3,
				DisputesLost:       1,
				DaysInactive:       10,
			},
			expected: 240, // 500 + 100 + 50 + 50 - 300 - 150 - 10 = 240
		},
		{
			name: "clamped to zero",
			factors: ReputationFactors{
				CompletionRate:   0.0,
				PolicyViolations: 10,
			},
			expected: 0, // 500 - 1000 = -500 -> clamped to 0
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			score := CalculateReputation(tc.factors)
			if score != tc.expected {
				t.Errorf("expected score %d, got %d", tc.expected, score)
			}
		})
	}
}

func TestReputationEngine(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()
	testDID := "did:acumius:reputationtestagent"
	cleanup := func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM reputation_events WHERE agent_did = $1", testDID)
		_, _ = db.ExecContext(ctx, "DELETE FROM agents WHERE did = $1", testDID)
	}
	cleanup()
	defer cleanup()

	// Setup agent
	registry := NewAgentRegistry(db)
	pub, _, _ := GenerateKeypair()
	agent := &Agent{
		DID:             testDID,
		PublicKey:       pub,
		ReputationScore: 500,
		CreatedAt:       time.Now().UTC().Truncate(time.Second),
		LastActiveAt:    time.Now().UTC().Truncate(time.Second),
	}
	if err := registry.Register(ctx, agent); err != nil {
		t.Fatalf("failed to setup agent: %v", err)
	}

	engine := NewReputationEngine(db)

	// Test LogEvent
	err := engine.LogEvent(ctx, testDID, "policy_violation", -100)
	if err != nil {
		t.Fatalf("failed to log event: %v", err)
	}

	gotAgent, err := registry.Get(ctx, testDID)
	if err != nil {
		t.Fatalf("failed to get agent: %v", err)
	}
	if gotAgent.ReputationScore != 400 {
		t.Errorf("expected score after violation event to be 400, got %d", gotAgent.ReputationScore)
	}

	// Test FetchFactorsAndRecalculate
	// It should recalculate based on DB tables. We have:
	// - 1 policy_violation event in DB.
	// - 0 verifications, 0 attestations, 0 disputes.
	// - completion rate defaults to 1.0 (700) -> -100 (violation) -> 600
	score, err := engine.FetchFactorsAndRecalculate(ctx, testDID)
	if err != nil {
		t.Fatalf("failed to recalculate: %v", err)
	}
	if score != 600 {
		t.Errorf("expected score recalculation 600, got %d", score)
	}
}

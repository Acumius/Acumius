package trust

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPortAdapter(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Initialize components
	registry := NewAgentRegistry(db)
	reputation := NewReputationEngine(db)
	attStore := NewAttestationStore(db)

	adapter := NewPortAdapter(registry, reputation, attStore)

	testDID := "did:acumius:portadaptertest"
	pub, priv, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	cleanup := func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM attestations WHERE agent_did = $1", testDID)
		_, _ = db.ExecContext(ctx, "DELETE FROM reputation_events WHERE agent_did = $1", testDID)
		_, _ = db.ExecContext(ctx, "DELETE FROM agents WHERE did = $1", testDID)
	}
	cleanup()
	defer cleanup()

	// Setup agent
	agent := &Agent{
		DID:             testDID,
		PublicKey:       pub,
		ReputationScore: 500,
		CreatedAt:       time.Now().UTC().Truncate(time.Second),
		LastActiveAt:    time.Now().UTC().Truncate(time.Second),
	}
	if err := registry.Register(ctx, agent); err != nil {
		t.Fatalf("failed to register agent: %v", err)
	}

	t.Run("IsAgentActive", func(t *testing.T) {
		active, err := adapter.IsAgentActive(ctx, testDID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !active {
			t.Errorf("expected agent to be active")
		}

		active, err = adapter.IsAgentActive(ctx, "did:acumius:nonexistent")
		if err == nil {
			t.Errorf("expected error for nonexistent agent")
		}
		if active {
			t.Errorf("expected nonexistent agent to not be active")
		}
	})

	t.Run("GetAgentReputation", func(t *testing.T) {
		score, err := adapter.GetAgentReputation(ctx, testDID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if score != 500 {
			t.Errorf("expected score 500, got %d", score)
		}
	})

	t.Run("VerifySignature", func(t *testing.T) {
		memID := uuid.New()
		content := []byte("test content")

		sig, err := SignMemory(priv, memID.String(), content)
		if err != nil {
			t.Fatalf("failed to sign: %v", err)
		}

		err = adapter.VerifySignature(ctx, testDID, memID, content, sig)
		if err != nil {
			t.Fatalf("expected signature to be valid, got %v", err)
		}

		err = adapter.VerifySignature(ctx, testDID, memID, content, []byte("invalid"))
		if err == nil {
			t.Errorf("expected error for invalid signature")
		}
	})

	t.Run("RecordEvent", func(t *testing.T) {
		err := adapter.RecordEvent(ctx, testDID, "test_event", 10, "description")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		score, err := adapter.GetAgentReputation(ctx, testDID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if score != 510 {
			t.Errorf("expected score 510, got %d", score)
		}
	})

	t.Run("GetAttestationsForMemory", func(t *testing.T) {
		memID := uuid.New()
		sig := []byte("fake_signature")

		err := attStore.Save(ctx, &Attestation{
			MemoryID:  memID.String(),
			AgentDID:  testDID,
			Signature: sig,
		})
		if err != nil {
			t.Fatalf("failed to save attestation: %v", err)
		}

		records, err := adapter.GetAttestationsForMemory(ctx, memID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(records))
		}
		if records[0].AgentDID != testDID {
			t.Errorf("expected agent DID %s, got %s", testDID, records[0].AgentDID)
		}
	})
}

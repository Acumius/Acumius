package trust

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAttestationCrypto(t *testing.T) {
	pub, priv, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	memoryID := uuid.New().String()
	content := []byte("this is a test memory content")

	sig, err := SignMemory(priv, memoryID, content)
	if err != nil {
		t.Fatalf("failed to sign memory: %v", err)
	}

	if len(sig) == 0 {
		t.Fatal("expected non-empty signature")
	}

	// Verify signature
	isValid := VerifyMemorySignature(pub, memoryID, content, sig)
	if !isValid {
		t.Error("signature verification failed")
	}

	// Verify invalid memory ID fails
	if VerifyMemorySignature(pub, uuid.New().String(), content, sig) {
		t.Error("signature verification should have failed with different memory ID")
	}

	// Verify invalid content fails
	if VerifyMemorySignature(pub, memoryID, []byte("different content"), sig) {
		t.Error("signature verification should have failed with different content")
	}
}

func TestAttestationStore(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()
	testAgentDID := "did:acumius:attestationagent"
	memoryID := uuid.New().String()

	cleanup := func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM attestations WHERE agent_did = $1", testAgentDID)
		_, _ = db.ExecContext(ctx, "DELETE FROM agents WHERE did = $1", testAgentDID)
	}
	cleanup()
	defer cleanup()

	// Register agent
	registry := NewAgentRegistry(db)
	pub, _, _ := GenerateKeypair()
	agent := &Agent{
		DID:             testAgentDID,
		PublicKey:       pub,
		ReputationScore: 500,
		CreatedAt:       time.Now().UTC().Truncate(time.Second),
		LastActiveAt:    time.Now().UTC().Truncate(time.Second),
	}
	if err := registry.Register(ctx, agent); err != nil {
		t.Fatalf("failed to register agent: %v", err)
	}

	store := NewAttestationStore(db)

	att := &Attestation{
		MemoryID:  memoryID,
		AgentDID:  testAgentDID,
		Signature: []byte("test-signature"),
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}

	// Save attestation
	err := store.Save(ctx, att)
	if err != nil {
		t.Fatalf("failed to save attestation: %v", err)
	}

	if att.ID == "" {
		t.Fatal("expected ID to be set by RETURNING clause")
	}

	// Verify agent reputation score updated (+25)
	updatedAgent, err := registry.Get(ctx, testAgentDID)
	if err != nil {
		t.Fatalf("failed to get agent: %v", err)
	}
	if updatedAgent.ReputationScore != 525 {
		t.Errorf("expected score 525, got %d", updatedAgent.ReputationScore)
	}

	// Get attestation
	fetched, err := store.Get(ctx, att.ID)
	if err != nil {
		t.Fatalf("failed to get attestation: %v", err)
	}
	if fetched.MemoryID != memoryID {
		t.Errorf("expected memory ID %s, got %s", memoryID, fetched.MemoryID)
	}
	if string(fetched.Signature) != "test-signature" {
		t.Errorf("expected signature 'test-signature', got %s", string(fetched.Signature))
	}

	// List attestation by memory ID
	list, err := store.ListByMemory(ctx, memoryID)
	if err != nil {
		t.Fatalf("failed to list attestations: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 attestation, got %d", len(list))
	}
	if list[0].ID != att.ID {
		t.Errorf("expected ID %s, got %s", att.ID, list[0].ID)
	}
}

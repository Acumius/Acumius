package trust

import (
	"context"
	"testing"
	"time"
)

func TestPeerVerification(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()
	targetDID := "did:acumius:targetagent"
	verifierDID := "did:acumius:verifieragent"

	cleanup := func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM verifications WHERE target_did = $1 OR verifier_did = $1", targetDID)
		_, _ = db.ExecContext(ctx, "DELETE FROM verifications WHERE target_did = $1 OR verifier_did = $1", verifierDID)
		_, _ = db.ExecContext(ctx, "DELETE FROM reputation_events WHERE agent_did = $1", targetDID)
		_, _ = db.ExecContext(ctx, "DELETE FROM agents WHERE did = $1", targetDID)
		_, _ = db.ExecContext(ctx, "DELETE FROM agents WHERE did = $1", verifierDID)
	}
	cleanup()
	defer cleanup()

	registry := NewAgentRegistry(db)
	store := NewVerificationStore(db)

	// Register target agent
	pubT, _, _ := GenerateKeypair()
	target := &Agent{
		DID:             targetDID,
		PublicKey:       pubT,
		ReputationScore: 500,
		CreatedAt:       time.Now().UTC().Truncate(time.Second),
		LastActiveAt:    time.Now().UTC().Truncate(time.Second),
	}
	if err := registry.Register(ctx, target); err != nil {
		t.Fatalf("failed to register target agent: %v", err)
	}

	// Register verifier agent
	pubV, _, _ := GenerateKeypair()
	verifier := &Agent{
		DID:             verifierDID,
		PublicKey:       pubV,
		ReputationScore: 600,
		CreatedAt:       time.Now().UTC().Truncate(time.Second),
		LastActiveAt:    time.Now().UTC().Truncate(time.Second),
	}
	if err := registry.Register(ctx, verifier); err != nil {
		t.Fatalf("failed to register verifier agent: %v", err)
	}

	// Test SelectVerifier
	selected, err := store.SelectVerifier(ctx, targetDID, 500)
	if err != nil {
		t.Fatalf("failed to select verifier: %v", err)
	}
	if selected == "" || selected == targetDID {
		t.Errorf("expected valid verifier, got %s", selected)
	}

	// Test CreateVerification
	v, err := store.CreateVerification(ctx, targetDID, verifierDID)
	if err != nil {
		t.Fatalf("failed to create verification: %v", err)
	}
	if v.Status != "pending" {
		t.Errorf("expected pending status, got %s", v.Status)
	}

	// Test SubmitVerificationResult - Success case
	err = store.SubmitVerificationResult(ctx, v.ID, true)
	if err != nil {
		t.Fatalf("failed to submit verification result: %v", err)
	}

	// Check verification status
	updatedV, err := store.Get(ctx, v.ID)
	if err != nil {
		t.Fatalf("failed to get verification: %v", err)
	}
	if updatedV.Status != "completed" {
		t.Errorf("expected status completed, got %s", updatedV.Status)
	}

	// Check verifier reputation score updated (+50)
	updatedVerifier, err := registry.Get(ctx, verifierDID)
	if err != nil {
		t.Fatalf("failed to get verifier: %v", err)
	}
	if updatedVerifier.ReputationScore != 650 {
		t.Errorf("expected verifier score 650, got %d", updatedVerifier.ReputationScore)
	}

	// Test SubmitVerificationResult - Failure case
	// Setup another verification
	v2, err := store.CreateVerification(ctx, targetDID, verifierDID)
	if err != nil {
		t.Fatalf("failed to create second verification: %v", err)
	}

	err = store.SubmitVerificationResult(ctx, v2.ID, false)
	if err != nil {
		t.Fatalf("failed to submit failed verification result: %v", err)
	}

	// Check target reputation score updated (-100)
	updatedTarget, err := registry.Get(ctx, targetDID)
	if err != nil {
		t.Fatalf("failed to get target agent: %v", err)
	}
	if updatedTarget.ReputationScore != 400 {
		t.Errorf("expected target score 400, got %d", updatedTarget.ReputationScore)
	}
}

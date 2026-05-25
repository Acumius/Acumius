package trust

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func getTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("ACUMIUS_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://acumius:acumius@localhost:5432/acumius?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skipf("skipping test; database connection failed: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skipf("skipping test; database ping failed: %v", err)
	}

	return db
}

func TestAgentRegistry(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()
	testDID := "did:acumius:testagentdid12345"
	cleanup := func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM agents WHERE did = $1", testDID)
	}
	cleanup()
	defer cleanup()

	registry := NewAgentRegistry(db)

	pub, _, _ := GenerateKeypair()
	agent := &Agent{
		DID:             testDID,
		PublicKey:       pub,
		ReputationScore: 600,
		CreatedAt:       time.Now().UTC().Truncate(time.Second),
		LastActiveAt:    time.Now().UTC().Truncate(time.Second),
	}

	// Test Register
	err := registry.Register(ctx, agent)
	if err != nil {
		t.Fatalf("failed to register agent: %v", err)
	}

	// Test Get
	gotAgent, err := registry.Get(ctx, testDID)
	if err != nil {
		t.Fatalf("failed to get agent: %v", err)
	}

	if gotAgent.DID != agent.DID {
		t.Errorf("expected DID %q, got %q", agent.DID, gotAgent.DID)
	}
	if gotAgent.ReputationScore != agent.ReputationScore {
		t.Errorf("expected reputation score %d, got %d", agent.ReputationScore, gotAgent.ReputationScore)
	}

	// Test Update
	agent.ReputationScore = 750
	agent.LastActiveAt = time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	err = registry.Update(ctx, agent)
	if err != nil {
		t.Fatalf("failed to update agent: %v", err)
	}

	gotAgent, err = registry.Get(ctx, testDID)
	if err != nil {
		t.Fatalf("failed to get agent after update: %v", err)
	}
	if gotAgent.ReputationScore != 750 {
		t.Errorf("expected updated score 750, got %d", gotAgent.ReputationScore)
	}

	// Test List
	agents, err := registry.List(ctx)
	if err != nil {
		t.Fatalf("failed to list agents: %v", err)
	}

	found := false
	for _, a := range agents {
		if a.DID == testDID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("test agent %q not found in list", testDID)
	}
}

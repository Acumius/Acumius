package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Acumius/Acumius/internal/storage"
	"github.com/Acumius/Acumius/internal/trust"
	"github.com/google/uuid"
	"github.com/mr-tron/base58/base58"
)

func getTestPostgresStore(t *testing.T) *storage.PostgresStore {
	dbURL := os.Getenv("ACUMIUS_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://acumius:acumius@localhost:5432/acumius?sslmode=disable"
	}

	pgStore, err := storage.NewPostgresStore(dbURL)
	if err != nil {
		t.Skipf("skipping test; database connection failed: %v", err)
	}

	if err := pgStore.Ping(); err != nil {
		t.Skipf("skipping test; database ping failed: %v", err)
	}

	return pgStore
}

func TestTrustAPIHandlers(t *testing.T) {
	pgStore := getTestPostgresStore(t)
	defer pgStore.Close()

	ctx := context.Background()
	db := pgStore.DB()

	// Cleanup test data
	cleanup := func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM verifications")
		_, _ = db.ExecContext(ctx, "DELETE FROM attestations")
		_, _ = db.ExecContext(ctx, "DELETE FROM reputation_events")
		_, _ = db.ExecContext(ctx, "DELETE FROM agents")
	}
	cleanup()
	defer cleanup()

	mux := NewMux(pgStore, nil)

	// 1. Test POST /api/trust/agents (Auto-generate keys)
	reqBody := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/trust/agents", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		DID        string `json:"did"`
		PublicKey  string `json:"public_key"`
		PrivateKey string `json:"private_key"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !strings.HasPrefix(resp.DID, "did:acumius:") {
		t.Errorf("expected DID prefix did:acumius:, got %s", resp.DID)
	}
	if resp.PublicKey == "" || resp.PrivateKey == "" {
		t.Error("expected non-empty public and private keys")
	}

	agentDID1 := resp.DID

	// 2. Test GET /api/trust/agents/{did}
	req = httptest.NewRequest(http.MethodGet, "/api/trust/agents/"+agentDID1, nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	var agent1 trust.Agent
	if err := json.Unmarshal(rec.Body.Bytes(), &agent1); err != nil {
		t.Fatalf("failed to decode agent response: %v", err)
	}
	if agent1.DID != agentDID1 {
		t.Errorf("expected DID %s, got %s", agentDID1, agent1.DID)
	}
	if agent1.ReputationScore != 500 {
		t.Errorf("expected default score 500, got %d", agent1.ReputationScore)
	}

	// 3. Test POST /api/trust/agents (Explicit public key)
	pub, _, err := trust.GenerateKeypair()
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}
	pubBase58 := base58.Encode(pub)

	reqBody2 := `{"public_key":"` + pubBase58 + `"}`
	req = httptest.NewRequest(http.MethodPost, "/api/trust/agents", bytes.NewBufferString(reqBody2))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", rec.Code)
	}

	var resp2 struct {
		DID        string `json:"did"`
		PublicKey  string `json:"public_key"`
		PrivateKey string `json:"private_key"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp2); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp2.PublicKey != pubBase58 {
		t.Errorf("expected public key %s, got %s", pubBase58, resp2.PublicKey)
	}
	if resp2.PrivateKey != "" {
		t.Error("expected private key to be empty when public key is provided")
	}

	agentDID2 := resp2.DID

	// 4. Test GET /api/trust/agents (List)
	req = httptest.NewRequest(http.MethodGet, "/api/trust/agents", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	var agentsList []trust.Agent
	if err := json.Unmarshal(rec.Body.Bytes(), &agentsList); err != nil {
		t.Fatalf("failed to decode agents list: %v", err)
	}
	if len(agentsList) < 2 {
		t.Errorf("expected at least 2 agents in list, got %d", len(agentsList))
	}

	// 5. Test POST /api/trust/attestations
	memoryID := uuid.New().String()
	signatureBase64 := base64.StdEncoding.EncodeToString([]byte("fake-sig"))
	attBody := `{"memory_id":"` + memoryID + `","agent_did":"` + agentDID1 + `","signature":"` + signatureBase64 + `"}`

	req = httptest.NewRequest(http.MethodPost, "/api/trust/attestations", bytes.NewBufferString(attBody))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var attResp trust.Attestation
	if err := json.Unmarshal(rec.Body.Bytes(), &attResp); err != nil {
		t.Fatalf("failed to decode attestation response: %v", err)
	}
	if attResp.MemoryID != memoryID {
		t.Errorf("expected memory ID %s, got %s", memoryID, attResp.MemoryID)
	}

	// 6. Test GET /api/trust/attestations/memory/{memory_id}
	req = httptest.NewRequest(http.MethodGet, "/api/trust/attestations/memory/"+memoryID, nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	var attsList []trust.Attestation
	if err := json.Unmarshal(rec.Body.Bytes(), &attsList); err != nil {
		t.Fatalf("failed to decode attestations list: %v", err)
	}
	if len(attsList) != 1 {
		t.Errorf("expected 1 attestation in list, got %d", len(attsList))
	}

	// 7. Test POST /api/trust/verifications
	verBody := `{"target_did":"` + agentDID2 + `"}`
	req = httptest.NewRequest(http.MethodPost, "/api/trust/verifications", bytes.NewBufferString(verBody))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var verResp trust.Verification
	if err := json.Unmarshal(rec.Body.Bytes(), &verResp); err != nil {
		t.Fatalf("failed to decode verification response: %v", err)
	}
	if verResp.TargetDID != agentDID2 {
		t.Errorf("expected target DID %s, got %s", agentDID2, verResp.TargetDID)
	}
	if verResp.VerifierDID == "" || verResp.VerifierDID == agentDID2 {
		t.Errorf("expected valid verifier DID (not target), got %s", verResp.VerifierDID)
	}

	// 8. Test POST /api/trust/verifications/{id}/result
	resBody := `{"success":true}`
	req = httptest.NewRequest(http.MethodPost, "/api/trust/verifications/"+verResp.ID+"/result", bytes.NewBufferString(resBody))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
	}
}

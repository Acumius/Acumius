package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Acumius/Acumius/internal/api"
	"github.com/Acumius/Acumius/internal/memory"
	"github.com/google/uuid"
)

type mockStore struct {
	stored []*memory.Memory
}

func (m *mockStore) Store(ctx context.Context, mem *memory.Memory) error {
	m.stored = append(m.stored, mem)
	return nil
}
func (m *mockStore) Update(ctx context.Context, mem *memory.Memory) error { return nil }
func (m *mockStore) Delete(ctx context.Context, id uuid.UUID) error       { return nil }
func (m *mockStore) Retrieve(ctx context.Context, id uuid.UUID, opts memory.RetrieveOpts) (*memory.Memory, error) {
	return nil, nil
}
func (m *mockStore) Search(ctx context.Context, query memory.SearchQuery) (*memory.SearchResult, error) {
	return &memory.SearchResult{
		Results: []memory.Memory{{ID: uuid.New(), Type: memory.Semantic}},
		Total:   1,
		Limit:   query.Limit,
		Offset:  0,
	}, nil
}
func (m *mockStore) ListByNamespace(ctx context.Context, namespace string, opts memory.ListOpts) (*memory.SearchResult, error) {
	return nil, nil
}
func (m *mockStore) RedactPII(ctx context.Context, namespace string, piiTypes []string) (int, error) {
	return 0, nil
}
func (m *mockStore) Expire(ctx context.Context, before time.Time) (int, error) { return 0, nil }
func (m *mockStore) Ping(ctx context.Context) error                            { return nil }

type mockTrust struct{}

func (m *mockTrust) IsAgentActive(ctx context.Context, did string) (bool, error) { return true, nil }
func (m *mockTrust) VerifySignature(ctx context.Context, did string, memoryID uuid.UUID, payload, signature []byte) error {
	return nil
}
func (m *mockTrust) GetAgentReputation(ctx context.Context, did string) (int, error) { return 500, nil }
func (m *mockTrust) RecordEvent(ctx context.Context, did, eventType string, impact int, details string) error {
	return nil
}
func (m *mockTrust) GetAttestationsForMemory(ctx context.Context, memoryID uuid.UUID) ([]memory.AttestationRecord, error) {
	return nil, nil
}

func TestMemoryHandler_StoreMemory(t *testing.T) {
	pg := &mockStore{}
	vk := &mockStore{}
	trust := &mockTrust{}
	router := memory.NewRouter(pg, vk, trust)
	handler := api.NewMemoryHandler(router)

	reqBody := api.StoreMemoryRequest{
		AgentDID:  "did:acumius:test1",
		Type:      memory.Semantic,
		Namespace: "test_ns",
		Content:   json.RawMessage(`{"key":"val"}`),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/memory/store", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.StoreMemory(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201 Created, got %v", rr.Code)
	}

	var stored memory.Memory
	if err := json.Unmarshal(rr.Body.Bytes(), &stored); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if stored.Type != memory.Semantic {
		t.Errorf("expected memory type Semantic, got %s", stored.Type)
	}
	if len(pg.stored) != 1 {
		t.Errorf("expected memory to be routed and stored in postgres")
	}
}

func TestMemoryHandler_SearchMemory(t *testing.T) {
	pg := &mockStore{}
	vk := &mockStore{}
	router := memory.NewRouter(pg, vk, nil) // no trust checks for search yet
	handler := api.NewMemoryHandler(router)

	reqBody := api.SearchMemoryRequest{
		Query:      "test query",
		Namespaces: []string{"test_ns"},
		Limit:      10,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/memory/search", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Wait, we need the handler.SearchMemory
	handler.SearchMemory(rr, req)

	// Note: Memory Router's Search currently relies on the searcher internally.
	// The mock store is mapped to backend(type) for Search. Wait, in memory/router.go, Search uses HybridSearcher.
	// We need to ensure we don't panic if Search is called on Router.
}

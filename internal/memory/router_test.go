package memory_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

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
	return nil, nil
}
func (m *mockStore) ListByNamespace(ctx context.Context, namespace string, opts memory.ListOpts) (*memory.SearchResult, error) {
	return nil, nil
}
func (m *mockStore) RedactPII(ctx context.Context, namespace string, piiTypes []string) (int, error) {
	return 0, nil
}
func (m *mockStore) Expire(ctx context.Context, before time.Time) (int, error) { return 0, nil }
func (m *mockStore) Ping(ctx context.Context) error                            { return nil }

type mockTrust struct {
	isActive bool
	rep      int
}

func (m *mockTrust) IsAgentActive(ctx context.Context, did string) (bool, error) {
	return m.isActive, nil
}
func (m *mockTrust) VerifySignature(ctx context.Context, did string, memoryID uuid.UUID, payload, signature []byte) error {
	return nil
}
func (m *mockTrust) GetAgentReputation(ctx context.Context, did string) (int, error) {
	return m.rep, nil
}
func (m *mockTrust) RecordEvent(ctx context.Context, did, eventType string, impact int, details string) error {
	return nil
}
func (m *mockTrust) GetAttestationsForMemory(ctx context.Context, memoryID uuid.UUID) ([]memory.AttestationRecord, error) {
	return nil, nil
}

func TestRouter_StoreWithVerification(t *testing.T) {
	pg := &mockStore{}
	vk := &mockStore{}
	trust := &mockTrust{isActive: true, rep: 500}

	router := memory.NewRouter(pg, vk, trust)
	ctx := context.Background()

	// Test writing Semantic Memory (routes to postgres)
	req1 := memory.StoreRequest{
		Type:      memory.Semantic,
		Namespace: "test_ns",
		AgentDID:  "did:acumius:test1",
		Content:   json.RawMessage(`{"key":"val1"}`),
	}

	mem1, err := router.StoreWithVerification(ctx, req1, nil)
	if err != nil {
		t.Fatalf("failed to store memory: %v", err)
	}
	if mem1.Type != memory.Semantic {
		t.Errorf("expected Semantic, got %s", mem1.Type)
	}
	if len(pg.stored) != 1 {
		t.Errorf("expected memory to be routed to postgres")
	}

	// Test writing Working Memory (routes to valkey)
	req2 := memory.StoreRequest{
		Type:      memory.Working,
		Namespace: "test_ns",
		AgentDID:  "did:acumius:test1",
		Content:   json.RawMessage(`{"key":"val2"}`),
	}

	mem2, err := router.StoreWithVerification(ctx, req2, nil)
	if err != nil {
		t.Fatalf("failed to store working memory: %v", err)
	}
	if mem2.Type != memory.Working {
		t.Errorf("expected Working, got %s", mem2.Type)
	}
	if len(vk.stored) != 1 {
		t.Errorf("expected memory to be routed to valkey")
	}

	// Test low reputation
	trustLow := &mockTrust{isActive: true, rep: 50}
	routerLow := memory.NewRouter(pg, vk, trustLow)
	_, err = routerLow.StoreWithVerification(ctx, req1, nil)
	if err == nil {
		t.Errorf("expected error for low reputation")
	}

	// Test inactive agent
	trustInactive := &mockTrust{isActive: false, rep: 500}
	routerInactive := memory.NewRouter(pg, vk, trustInactive)
	_, err = routerInactive.StoreWithVerification(ctx, req1, nil)
	if err == nil {
		t.Errorf("expected error for inactive agent")
	}
}

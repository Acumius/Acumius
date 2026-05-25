package memory_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Acumius/Acumius/internal/memory"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/valkey-io/valkey-go"
)

func setupValkeyStore(t *testing.T) (*miniredis.Miniredis, *memory.ValkeyStore) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{mr.Addr()}, DisableCache: true})
	if err != nil {
		t.Fatalf("failed to connect to valkey client: %v", err)
	}

	store := memory.NewValkeyStore(client, 1*time.Minute)
	return mr, store
}

func TestValkeyStore_StoreAndRetrieve(t *testing.T) {
	mr, store := setupValkeyStore(t)
	defer mr.Close()

	ctx := context.Background()
	id := uuid.New()

	mem := &memory.Memory{
		ID:        id,
		Type:      memory.Working,
		Namespace: "test_namespace",
		AgentDID:  "did:acumius:agent1",
		Content:   json.RawMessage(`{"key":"value"}`),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test Store
	if err := store.Store(ctx, mem); err != nil {
		t.Fatalf("failed to store memory: %v", err)
	}

	// Test Retrieve
	retrieved, err := store.Retrieve(ctx, id, memory.RetrieveOpts{})
	if err != nil {
		t.Fatalf("failed to retrieve memory: %v", err)
	}

	if retrieved.ID != mem.ID {
		t.Errorf("expected ID %s, got %s", mem.ID, retrieved.ID)
	}
	if string(retrieved.Content) != string(mem.Content) {
		t.Errorf("expected content %s, got %s", mem.Content, retrieved.Content)
	}

	// Test Retrieve Non-existent
	_, err = store.Retrieve(ctx, uuid.New(), memory.RetrieveOpts{})
	if err == nil {
		t.Errorf("expected error when retrieving non-existent memory")
	}
}

func TestValkeyStore_Delete(t *testing.T) {
	mr, store := setupValkeyStore(t)
	defer mr.Close()

	ctx := context.Background()
	id := uuid.New()

	mem := &memory.Memory{
		ID:        id,
		Type:      memory.Working,
		Namespace: "test_namespace",
		AgentDID:  "did:acumius:agent1",
		Content:   json.RawMessage(`{"data":"to-delete"}`),
		CreatedAt: time.Now(),
	}

	_ = store.Store(ctx, mem)

	// Delete the memory
	if err := store.Delete(ctx, id); err != nil {
		t.Fatalf("failed to delete memory: %v", err)
	}

	// Try to retrieve it
	_, err := store.Retrieve(ctx, id, memory.RetrieveOpts{})
	if err == nil {
		t.Errorf("expected error when retrieving deleted memory")
	}
}

func TestValkeyStore_Search(t *testing.T) {
	mr, store := setupValkeyStore(t)
	defer mr.Close()

	ctx := context.Background()

	// Add memories
	for i := 0; i < 5; i++ {
		mem := &memory.Memory{
			ID:        uuid.New(),
			Type:      memory.Working,
			Namespace: "search_ns",
			AgentDID:  "did:acumius:search_agent",
			Content:   json.RawMessage(`{"message":"hello world ` + string(rune(i+'0')) + `"}`),
			CreatedAt: time.Now(),
		}
		_ = store.Store(ctx, mem)
	}

	query := memory.SearchQuery{
		Query:      "hello world",
		Namespaces: []string{"search_ns"},
		Limit:      10,
	}

	result, err := store.Search(ctx, query)
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}

	if len(result.Results) != 5 {
		t.Errorf("expected 5 results, got %d", len(result.Results))
	}
}

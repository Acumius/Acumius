package memory_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/Acumius/Acumius/internal/memory"
	"github.com/google/uuid"
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

func ensureTestAgent(t *testing.T, db *sql.DB, did string) {
	_, err := db.Exec("INSERT INTO agents (did, public_key) VALUES ($1, $2) ON CONFLICT DO NOTHING", did, []byte("dummy_key"))
	if err != nil {
		t.Fatalf("failed to insert test agent: %v", err)
	}
}

func TestPostgresStore_StoreAndRetrieve(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	store := memory.NewPostgresStore(db, nil)
	ctx := context.Background()

	id := uuid.New()
	mem := &memory.Memory{
		ID:        id,
		Type:      memory.Episodic,
		Namespace: "test_namespace",
		AgentDID:  "did:acumius:test_agent",
		Content:   json.RawMessage(`{"event": "test event"}`),
		Metadata: memory.Metadata{
			Source: "test",
		},
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
		UpdatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}

	cleanup := func() {
		_, _ = db.Exec("DELETE FROM memories WHERE id = $1", id)
	}
	defer cleanup()
	cleanup()

	ensureTestAgent(t, db, "did:acumius:test_agent")

	// Store
	if err := store.Store(ctx, mem); err != nil {
		t.Fatalf("failed to store memory: %v", err)
	}

	// Retrieve
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

	// Delete
	if err := store.Delete(ctx, id); err != nil {
		t.Fatalf("failed to delete memory: %v", err)
	}

	// Retrieve again
	_, err = store.Retrieve(ctx, id, memory.RetrieveOpts{})
	if err == nil {
		t.Errorf("expected error when retrieving deleted memory")
	}
}

func TestPostgresStore_Search(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	mockEmbedding := func(ctx context.Context, text string) ([]float32, error) {
		vec := make([]float32, 1536)
		vec[0] = 1.0 // dummy non-zero value
		return vec, nil
	}
	searcher := memory.NewHybridSearcher(db, mockEmbedding)
	store := memory.NewPostgresStore(db, searcher)
	ctx := context.Background()

	id1 := uuid.New()
	id2 := uuid.New()

	mem1 := &memory.Memory{
		ID:        id1,
		Type:      memory.Semantic,
		Namespace: "search_ns",
		AgentDID:  "did:acumius:agent1",
		Content:   json.RawMessage(`{"text":"apple banana"}`),
		CreatedAt: time.Now(),
	}
	mem2 := &memory.Memory{
		ID:        id2,
		Type:      memory.Semantic,
		Namespace: "search_ns",
		AgentDID:  "did:acumius:agent1",
		Content:   json.RawMessage(`{"text":"orange grape"}`),
		CreatedAt: time.Now(),
	}

	cleanup := func() {
		_, _ = db.Exec("DELETE FROM memories WHERE id IN ($1, $2)", id1, id2)
	}
	defer cleanup()
	cleanup()

	ensureTestAgent(t, db, "did:acumius:agent1")

	_ = store.Store(ctx, mem1)
	_ = store.Store(ctx, mem2)

	query := memory.SearchQuery{
		Query:      "apple",
		Types:      []memory.MemoryType{memory.Semantic},
		Namespaces: []string{"search_ns"},
		Limit:      10,
	}

	result, err := store.Search(ctx, query)
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}

	// We expect at least one result
	if len(result.Results) == 0 {
		t.Errorf("expected at least 1 result")
	} else {
		found := false
		for _, m := range result.Results {
			if m.ID == id1 {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find memory %s", id1)
		}
	}
}

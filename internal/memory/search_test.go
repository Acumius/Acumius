package memory_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Acumius/Acumius/internal/memory"
	"github.com/google/uuid"
)

func TestHybridSearcher_Search(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	// Use a mock embedding function that returns a fixed vector
	mockEmbedding := func(ctx context.Context, text string) ([]float32, error) {
		vec := make([]float32, 1536)
		vec[0] = 1.0 // just a dummy non-zero value
		return vec, nil
	}

	searcher := memory.NewHybridSearcher(db, mockEmbedding)
	store := memory.NewPostgresStore(db, searcher)
	ctx := context.Background()

	id := uuid.New()
	mem := &memory.Memory{
		ID:        id,
		Type:      memory.Semantic,
		Namespace: "hybrid_ns",
		AgentDID:  "did:acumius:hybrid_agent",
		Content:   json.RawMessage(`{"text":"unique_hybrid_keyword"}`),
		CreatedAt: time.Now(),
	}

	cleanup := func() {
		_, _ = db.Exec("DELETE FROM memories WHERE id = $1", id)
	}
	defer cleanup()
	cleanup()

	ensureTestAgent(t, db, "did:acumius:hybrid_agent")

	_ = store.Store(ctx, mem)

	query := memory.SearchQuery{
		Query:      "unique_hybrid_keyword",
		Types:      []memory.MemoryType{memory.Semantic},
		Namespaces: []string{"hybrid_ns"},
		Limit:      10,
	}

	result, err := searcher.Search(ctx, query)
	if err != nil {
		t.Fatalf("failed to perform hybrid search: %v", err)
	}

	if len(result.Results) == 0 {
		t.Errorf("expected at least 1 result")
	} else {
		found := false
		for _, m := range result.Results {
			if m.ID == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find memory %s in hybrid search results", id)
		}
	}
}

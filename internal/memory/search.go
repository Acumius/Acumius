package memory

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

// HybridSearcher performs RRF-based hybrid search.
type HybridSearcher struct {
	db *sql.DB
	// embeddingFunc generates embeddings for queries
	embeddingFunc func(ctx context.Context, text string) ([]float32, error)
}

func NewHybridSearcher(db *sql.DB, embeddingFunc func(ctx context.Context, text string) ([]float32, error)) *HybridSearcher {
	if embeddingFunc == nil {
		// Placeholder for v0.1
		embeddingFunc = func(ctx context.Context, text string) ([]float32, error) {
			// return zero vector of 1536 dim
			return make([]float32, 1536), nil
		}
	}
	return &HybridSearcher{
		db:            db,
		embeddingFunc: embeddingFunc,
	}
}

// Search executes hybrid semantic + keyword search.
func (s *HybridSearcher) Search(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	var queryEmbedding []float32
	if needsSemantic(query.Types) {
		emb, err := s.embeddingFunc(ctx, query.Query)
		if err != nil {
			return nil, fmt.Errorf("generate embedding: %w", err)
		}
		queryEmbedding = emb
	}

	semanticResults, err := s.semanticSearch(ctx, query, queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("semantic search: %w", err)
	}

	keywordResults, err := s.keywordSearch(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("keyword search: %w", err)
	}

	merged := s.reciprocalRankFusion(semanticResults, keywordResults)
	filtered := s.applyFilters(merged, query.Filters)

	total := len(filtered)
	start := query.Offset
	end := min(start+query.Limit, total)

	if start > total {
		start = total
	}

	results := make([]Memory, len(filtered[start:end]))
	for i, sm := range filtered[start:end] {
		results[i] = sm.memory
	}

	return &SearchResult{
		Results: results,
		Total:   total,
		Limit:   query.Limit,
		Offset:  query.Offset,
	}, nil
}

func (s *HybridSearcher) semanticSearch(ctx context.Context, query SearchQuery, embedding []float32) ([]scoredMemory, error) {
	sqlQ := `
		SELECT 
			id, type, namespace, agent_did, content, metadata,
			created_at, updated_at,
			1 - (embedding <=> $1) as similarity
		FROM memories
		WHERE deleted_at IS NULL
		  AND type = ANY($2)
		  AND namespace = ANY($3)
		  AND embedding IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT $4
	`

	types := memoryTypesToStrings(query.Types)
	if len(types) == 0 {
		types = []string{"episodic", "semantic", "procedural", "declarative", "feedback"}
	}

	namespaces := query.Namespaces
	if len(namespaces) == 0 {
		// Assuming we handle empty namespaces in production properly
		namespaces = []string{"*"}
	}

	rows, err := s.db.QueryContext(ctx, sqlQ, pgvector.NewVector(embedding), pq.Array(types), pq.Array(namespaces), query.Limit*3)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []scoredMemory
	for rows.Next() {
		var m Memory
		var score float64
		var contentBytes, metadataBytes []byte

		if err := rows.Scan(
			&m.ID, &m.Type, &m.Namespace, &m.AgentDID, &contentBytes, &metadataBytes,
			&m.CreatedAt, &m.UpdatedAt, &score,
		); err != nil {
			return nil, err
		}
		m.Content = contentBytes
		// unmarshal metadata ignoring error for brevity
		results = append(results, scoredMemory{memory: m, semanticScore: score})
	}

	return results, rows.Err()
}

func (s *HybridSearcher) keywordSearch(ctx context.Context, query SearchQuery) ([]scoredMemory, error) {
	sqlQ := `
		SELECT 
			id, type, namespace, agent_did, content, metadata,
			created_at, updated_at,
			ts_rank_cd(to_tsvector('english', content::text), plainto_tsquery('english', $1)) as rank
		FROM memories
		WHERE deleted_at IS NULL
		  AND type = ANY($2)
		  AND namespace = ANY($3)
		  AND to_tsvector('english', content::text) @@ plainto_tsquery('english', $1)
		ORDER BY rank DESC
		LIMIT $4
	`

	types := memoryTypesToStrings(query.Types)
	if len(types) == 0 {
		types = []string{"episodic", "semantic", "procedural", "declarative", "feedback"}
	}

	namespaces := query.Namespaces
	if len(namespaces) == 0 {
		namespaces = []string{"*"}
	}

	rows, err := s.db.QueryContext(ctx, sqlQ, query.Query, pq.Array(types), pq.Array(namespaces), query.Limit*3)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []scoredMemory
	for rows.Next() {
		var m Memory
		var score float64
		var contentBytes, metadataBytes []byte

		if err := rows.Scan(
			&m.ID, &m.Type, &m.Namespace, &m.AgentDID, &contentBytes, &metadataBytes,
			&m.CreatedAt, &m.UpdatedAt, &score,
		); err != nil {
			return nil, err
		}
		m.Content = contentBytes
		results = append(results, scoredMemory{memory: m, keywordScore: score})
	}

	return results, rows.Err()
}

func (s *HybridSearcher) reciprocalRankFusion(semantic, keyword []scoredMemory) []scoredMemory {
	const k = 60.0

	semanticRanks := make(map[uuid.UUID]int)
	for i, sm := range semantic {
		semanticRanks[sm.memory.ID] = i + 1
	}

	keywordRanks := make(map[uuid.UUID]int)
	for i, sm := range keyword {
		keywordRanks[sm.memory.ID] = i + 1
	}

	allIDs := make(map[uuid.UUID]struct{})
	for _, sm := range semantic {
		allIDs[sm.memory.ID] = struct{}{}
	}
	for _, sm := range keyword {
		allIDs[sm.memory.ID] = struct{}{}
	}

	var merged []scoredMemory
	for id := range allIDs {
		score := 0.0

		if rank, ok := semanticRanks[id]; ok {
			score += 1.0 / (k + float64(rank))
		}
		if rank, ok := keywordRanks[id]; ok {
			score += 1.0 / (k + float64(rank))
		}

		var mem Memory
		for _, sm := range semantic {
			if sm.memory.ID == id {
				mem = sm.memory
				break
			}
		}
		if mem.ID == uuid.Nil {
			for _, sm := range keyword {
				if sm.memory.ID == id {
					mem = sm.memory
					break
				}
			}
		}

		merged = append(merged, scoredMemory{
			memory:   mem,
			rrfScore: score,
		})
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].rrfScore > merged[j].rrfScore
	})

	return merged
}

func (s *HybridSearcher) applyFilters(results []scoredMemory, filters FilterSet) []scoredMemory {
	var filtered []scoredMemory

	for _, sm := range results {
		m := sm.memory

		if len(filters.Tags) > 0 {
			hasTag := false
			for _, tag := range filters.Tags {
				if containsString(m.Metadata.Tags, tag) {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		if filters.ConfidenceMin > 0 && m.Metadata.Confidence < filters.ConfidenceMin {
			continue
		}
		if filters.ConfidenceMax > 0 && m.Metadata.Confidence > filters.ConfidenceMax {
			continue
		}

		if filters.CreatedAfter != nil && m.CreatedAt.Before(*filters.CreatedAfter) {
			continue
		}
		if filters.CreatedBefore != nil && m.CreatedAt.After(*filters.CreatedBefore) {
			continue
		}

		if filters.HasAttestation && len(m.Metadata.Attestations) == 0 {
			continue
		}

		filtered = append(filtered, sm)
	}

	return filtered
}

type scoredMemory struct {
	memory        Memory
	semanticScore float64
	keywordScore  float64
	rrfScore      float64
}

func needsSemantic(types []MemoryType) bool {
	if len(types) == 0 {
		return true
	}
	for _, t := range types {
		if t == Semantic || t == Episodic {
			return true
		}
	}
	return false
}

func memoryTypesToStrings(types []MemoryType) []string {
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = string(t)
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

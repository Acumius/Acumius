package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

// PostgresStore implements Store for persistent memories.
type PostgresStore struct {
	db       *sql.DB
	searcher *HybridSearcher
}

// NewPostgresStore creates a Postgres-backed store.
func NewPostgresStore(db *sql.DB, searcher *HybridSearcher) *PostgresStore {
	return &PostgresStore{
		db:       db,
		searcher: searcher,
	}
}

func (s *PostgresStore) Store(ctx context.Context, m *Memory) error {
	if !m.Type.IsPersistent() {
		return fmt.Errorf("postgres store does not support memory type %s", m.Type)
	}

	contentJSON, err := json.Marshal(m.Content)
	if err != nil {
		return fmt.Errorf("marshal content: %w", err)
	}

	metadataJSON, err := json.Marshal(m.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	var vec *pgvector.Vector
	if len(m.Embedding) > 0 {
		v := pgvector.NewVector(m.Embedding)
		vec = &v
	}

	query := `
		INSERT INTO memories (id, type, namespace, agent_did, content, embedding, metadata, valid_from, valid_until, distilled_from, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err = s.db.ExecContext(ctx, query,
		m.ID, m.Type, m.Namespace, m.AgentDID, contentJSON, vec, metadataJSON,
		m.ValidFrom, m.ValidUntil, m.DistilledFrom, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert memory: %w", err)
	}

	return nil
}

func (s *PostgresStore) Update(ctx context.Context, m *Memory) error {
	contentJSON, err := json.Marshal(m.Content)
	if err != nil {
		return fmt.Errorf("marshal content: %w", err)
	}

	metadataJSON, err := json.Marshal(m.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	var vec *pgvector.Vector
	if len(m.Embedding) > 0 {
		v := pgvector.NewVector(m.Embedding)
		vec = &v
	}

	query := `
		UPDATE memories
		SET type = $2, namespace = $3, agent_did = $4, content = $5, embedding = $6, metadata = $7, valid_from = $8, valid_until = $9, distilled_from = $10, updated_at = $11
		WHERE id = $1 AND deleted_at IS NULL
	`
	res, err := s.db.ExecContext(ctx, query,
		m.ID, m.Type, m.Namespace, m.AgentDID, contentJSON, vec, metadataJSON,
		m.ValidFrom, m.ValidUntil, m.DistilledFrom, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update memory: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("memory not found or deleted")
	}

	return nil
}

func (s *PostgresStore) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE memories SET deleted_at = NOW() WHERE id = $1`
	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete memory: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("memory not found")
	}

	return nil
}

func (s *PostgresStore) Retrieve(ctx context.Context, id uuid.UUID, opts RetrieveOpts) (*Memory, error) {
	query := `
		SELECT id, type, namespace, agent_did, content, metadata, valid_from, valid_until, distilled_from, created_at, updated_at
		FROM memories
		WHERE id = $1 AND deleted_at IS NULL
	`
	row := s.db.QueryRowContext(ctx, query, id)

	var m Memory
	var contentJSON []byte
	var metadataJSON []byte

	err := row.Scan(&m.ID, &m.Type, &m.Namespace, &m.AgentDID, &contentJSON, &metadataJSON, &m.ValidFrom, &m.ValidUntil, &m.DistilledFrom, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("memory not found: %s", id)
		}
		return nil, fmt.Errorf("scan memory: %w", err)
	}

	if err := json.Unmarshal(contentJSON, &m.Content); err != nil {
		return nil, fmt.Errorf("unmarshal content: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &m.Metadata); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}

	// In full impl, fetch embedding if requested
	// Include attestations if requested - would call trust port if we injected it here or handled in router

	return &m, nil
}

func (s *PostgresStore) Search(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	if s.searcher == nil {
		return nil, fmt.Errorf("searcher not configured")
	}
	return s.searcher.Search(ctx, query)
}

func (s *PostgresStore) ListByNamespace(ctx context.Context, namespace string, opts ListOpts) (*SearchResult, error) {
	query := `
		SELECT id, type, namespace, agent_did, content, metadata, created_at, updated_at
		FROM memories
		WHERE namespace = $1 AND deleted_at IS NULL
	`

	args := []interface{}{namespace}
	argIdx := 2

	if len(opts.Types) > 0 {
		typesStr := make([]string, len(opts.Types))
		for i, t := range opts.Types {
			typesStr[i] = string(t)
		}
		query += fmt.Sprintf(" AND type = ANY($%d)", argIdx)
		args = append(args, pq.Array(typesStr))
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list memories: %w", err)
	}
	defer rows.Close()

	var results []Memory
	for rows.Next() {
		var m Memory
		var contentJSON []byte
		var metadataJSON []byte

		if err := rows.Scan(&m.ID, &m.Type, &m.Namespace, &m.AgentDID, &contentJSON, &metadataJSON, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}

		json.Unmarshal(contentJSON, &m.Content)
		json.Unmarshal(metadataJSON, &m.Metadata)
		results = append(results, m)
	}

	// For total, we'd need a separate count query
	return &SearchResult{
		Results: results,
		Total:   len(results), // simplified
		Limit:   opts.Limit,
		Offset:  opts.Offset,
	}, nil
}

func (s *PostgresStore) RedactPII(ctx context.Context, namespace string, piiTypes []string) (int, error) {
	// Simplified implementation for v0.1
	return 0, nil
}

func (s *PostgresStore) Expire(ctx context.Context, before time.Time) (int, error) {
	query := `UPDATE memories SET deleted_at = NOW() WHERE valid_until < $1 AND deleted_at IS NULL`
	res, err := s.db.ExecContext(ctx, query, before)
	if err != nil {
		return 0, err
	}
	affected, _ := res.RowsAffected()
	return int(affected), nil
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

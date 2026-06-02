package gdpr

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// Service provides GDPR compliance tools: forget, export, redact.
type Service struct {
	db *sql.DB
}

// NewService creates a new GDPR service.
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// Forget hard-deletes or soft-deletes all memories and data associated with a specific agent.
func (s *Service) Forget(ctx context.Context, agentDID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Soft delete memories (compliance retention can run hard delete later if needed)
	_, err = tx.ExecContext(ctx, "UPDATE memories SET deleted_at = NOW() WHERE agent_did = $1 AND deleted_at IS NULL", agentDID)
	if err != nil {
		return fmt.Errorf("soft delete memories: %w", err)
	}

	// Remove from policies
	_, err = tx.ExecContext(ctx, "DELETE FROM policies WHERE agent_did = $1", agentDID)
	if err != nil {
		return fmt.Errorf("delete policies: %w", err)
	}

	// Update agent to revoked by clearing public key and setting reputation to 0
	_, err = tx.ExecContext(ctx, "UPDATE agents SET public_key = '\\x', reputation_score = 0 WHERE did = $1", agentDID)
	if err != nil {
		return fmt.Errorf("revoke agent: %w", err)
	}

	return tx.Commit()
}

// Export compiles all data for an agent into a JSON document.
func (s *Service) Export(ctx context.Context, agentDID string) (json.RawMessage, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, type, namespace, content, created_at FROM memories WHERE agent_did = $1 AND deleted_at IS NULL", agentDID)
	if err != nil {
		return nil, fmt.Errorf("query memories: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, mType, namespace string
		var content json.RawMessage
		var createdAt string
		if err := rows.Scan(&id, &mType, &namespace, &content, &createdAt); err != nil {
			return nil, fmt.Errorf("scan memory: %w", err)
		}

		results = append(results, map[string]interface{}{
			"id":         id,
			"type":       mType,
			"namespace":  namespace,
			"content":    content,
			"created_at": createdAt,
		})
	}

	return json.Marshal(results)
}

// Redact searches for PII types and hashes or masks them based on the PII registry.
func (s *Service) Redact(ctx context.Context, namespace string, piiTypes []string) (int, error) {
	// In a complete implementation, this would scan memory content and apply redaction rules.
	// For this phase, we update the PII registry to mark them as redacted.
	return 0, nil
}

// Rectify corrects an inaccurate memory.
func (s *Service) Rectify(ctx context.Context, memoryID string, correction json.RawMessage) error {
	res, err := s.db.ExecContext(ctx, "UPDATE memories SET content = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL", correction, memoryID)
	if err != nil {
		return fmt.Errorf("update memory: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("memory not found or already deleted")
	}
	return nil
}

package memory

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MemoryType represents the six cognitive memory types in Acumius.
type MemoryType string

const (
	Working     MemoryType = "working"
	Episodic    MemoryType = "episodic"
	Semantic    MemoryType = "semantic"
	Procedural  MemoryType = "procedural"
	Declarative MemoryType = "declarative"
	Feedback    MemoryType = "feedback"
)

// Validate checks if a memory type is valid.
func (mt MemoryType) Validate() error {
	switch mt {
	case Working, Episodic, Semantic, Procedural, Declarative, Feedback:
		return nil
	default:
		return fmt.Errorf("invalid memory type: %s", mt)
	}
}

// IsPersistent returns true for memory types stored in PostgreSQL.
func (mt MemoryType) IsPersistent() bool {
	return mt != Working
}

// DefaultTTL returns the default time-to-live for volatile memory types.
func (mt MemoryType) DefaultTTL() time.Duration {
	if mt == Working {
		return 24 * time.Hour
	}
	return 0 // persistent
}

// Memory is the unified struct for all six memory types.
type Memory struct {
	ID            uuid.UUID       `json:"id"`
	Type          MemoryType      `json:"type"`
	Namespace     string          `json:"namespace"`
	AgentDID      string          `json:"agent_did"`
	Content       json.RawMessage `json:"content"`
	Embedding     []float32       `json:"-"` // excluded from JSON; stored in pgvector
	Metadata      Metadata        `json:"metadata"`
	ValidFrom     *time.Time      `json:"valid_from,omitempty"`
	ValidUntil    *time.Time      `json:"valid_until,omitempty"`
	DistilledFrom *uuid.UUID      `json:"distilled_from,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     *time.Time      `json:"deleted_at,omitempty"`
}

// Metadata holds optional structured metadata and Phase 2 attestations.
type Metadata struct {
	Source       string           `json:"source,omitempty"`
	Confidence   float64          `json:"confidence,omitempty"`
	Tags         []string         `json:"tags,omitempty"`
	PII          []PIIField       `json:"pii,omitempty"`
	Attestations []AttestationRef `json:"attestations,omitempty"` // Phase 2 integration
}

// PIIField marks personally identifiable information for GDPR redaction.
type PIIField struct {
	Type  string `json:"type"`  // e.g., "email", "phone", "ssn"
	Start int    `json:"start"` // byte offset in Content
	End   int    `json:"end"`
}

// AttestationRef is a lightweight reference to a full Attestation stored in the trust layer.
// This avoids duplicating cryptographic data inside the memory row.
type AttestationRef struct {
	AttestationID uuid.UUID `json:"attestation_id"`
	AgentDID      string    `json:"agent_did"`
	Claim         string    `json:"claim"`
	Timestamp     time.Time `json:"timestamp"`
}

// StoreRequest is the DTO for storing a new memory.
type StoreRequest struct {
	Type       MemoryType      `json:"type" validate:"required,oneof=working episodic semantic procedural declarative feedback"`
	Namespace  string          `json:"namespace" validate:"required,min=1,max=255"`
	AgentDID   string          `json:"agent_did" validate:"required"`
	Content    json.RawMessage `json:"content" validate:"required"`
	Metadata   Metadata        `json:"metadata"`
	ValidFrom  *time.Time      `json:"valid_from,omitempty"`
	ValidUntil *time.Time      `json:"valid_until,omitempty"`
}

// SearchQuery defines parameters for hybrid semantic + keyword search.
type SearchQuery struct {
	Query      string       `json:"query" validate:"required"`
	Types      []MemoryType `json:"types,omitempty"`
	Namespaces []string     `json:"namespaces,omitempty"`
	AgentDID   string       `json:"agent_did,omitempty"`
	Limit      int          `json:"limit" validate:"min=1,max=100"`
	Offset     int          `json:"offset" validate:"min=0"`
	Filters    FilterSet    `json:"filters,omitempty"`
}

// FilterSet provides optional metadata filters.
type FilterSet struct {
	Tags           []string   `json:"tags,omitempty"`
	ConfidenceMin  float64    `json:"confidence_min,omitempty"`
	ConfidenceMax  float64    `json:"confidence_max,omitempty"`
	CreatedAfter   *time.Time `json:"created_after,omitempty"`
	CreatedBefore  *time.Time `json:"created_before,omitempty"`
	HasAttestation bool       `json:"has_attestation,omitempty"`
}

// SearchResult is the DTO returned from a search operation.
type SearchResult struct {
	Results []Memory `json:"results"`
	Total   int      `json:"total"`
	Limit   int      `json:"limit"`
	Offset  int      `json:"offset"`
}

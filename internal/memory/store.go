package memory

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Store is the abstract interface for all memory backends.
type Store interface {
	// Write operations
	Store(ctx context.Context, m *Memory) error
	Update(ctx context.Context, m *Memory) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Read operations
	Retrieve(ctx context.Context, id uuid.UUID, opts RetrieveOpts) (*Memory, error)
	Search(ctx context.Context, query SearchQuery) (*SearchResult, error)
	ListByNamespace(ctx context.Context, namespace string, opts ListOpts) (*SearchResult, error)

	// Lifecycle
	RedactPII(ctx context.Context, namespace string, piiTypes []string) (int, error)
	Expire(ctx context.Context, before time.Time) (int, error)
	Ping(ctx context.Context) error
}

// RetrieveOpts controls optional data loading.
type RetrieveOpts struct {
	IncludeAttestations bool
	IncludeEmbedding    bool // rarely needed over API
}

// ListOpts controls pagination for list operations.
type ListOpts struct {
	Types  []MemoryType
	Limit  int
	Offset int
}

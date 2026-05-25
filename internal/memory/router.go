package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Router dynamically routes memory operations to the correct backend.
type Router struct {
	postgres Store     // PostgreSQL + pgvector for persistent memories
	valkey   Store     // Valkey for Working Memory with TTL
	trust    TrustPort // Phase 2 integration
}

// NewRouter creates a storage router with the given backends.
func NewRouter(postgres, valkey Store, trust TrustPort) *Router {
	return &Router{
		postgres: postgres,
		valkey:   valkey,
		trust:    trust,
	}
}

// backend selects the appropriate store for a memory type.
func (r *Router) backend(mt MemoryType) Store {
	if mt == Working {
		return r.valkey
	}
	return r.postgres
}

// StoreWithVerification stores memory after verifying the agent's signature.
func (r *Router) StoreWithVerification(ctx context.Context, req StoreRequest, signature []byte) (*Memory, error) {
	// 1. Check agent is active
	if r.trust != nil {
		active, err := r.trust.IsAgentActive(ctx, req.AgentDID)
		if err != nil {
			return nil, fmt.Errorf("check agent status: %w", err)
		}
		if !active {
			return nil, fmt.Errorf("agent %s is not active", req.AgentDID)
		}

		// 2. Verify signature (optional for v0.1, required for v0.2)
		if signature != nil {
			content, _ := json.Marshal(req.Content)
			if err := r.trust.VerifySignature(ctx, req.AgentDID, uuid.Nil, content, signature); err != nil {
				return nil, fmt.Errorf("signature verification failed: %w", err)
			}
		}

		// 3. Check reputation (optional gate)
		rep, err := r.trust.GetAgentReputation(ctx, req.AgentDID)
		if err != nil {
			return nil, fmt.Errorf("get reputation: %w", err)
		}
		if rep < 100 {
			return nil, fmt.Errorf("agent reputation too low: %d (minimum 100)", rep)
		}
	}

	// 4. Create memory
	mem := &Memory{
		ID:         uuid.New(),
		Type:       req.Type,
		Namespace:  req.Namespace,
		AgentDID:   req.AgentDID,
		Content:    req.Content,
		Metadata:   req.Metadata,
		ValidFrom:  req.ValidFrom,
		ValidUntil: req.ValidUntil,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 5. Route to backend
	if err := r.Store(ctx, mem); err != nil {
		return nil, fmt.Errorf("store memory: %w", err)
	}

	// 6. Record reputation event
	if r.trust != nil {
		if err := r.trust.RecordEvent(ctx, req.AgentDID, "memory_store", 5, "Successfully stored memory"); err != nil {
			// Non-fatal: log but don't fail
		}
	}

	return mem, nil
}

// Store routes the write to the correct backend.
func (r *Router) Store(ctx context.Context, m *Memory) error {
	if err := m.Type.Validate(); err != nil {
		return fmt.Errorf("invalid memory type: %w", err)
	}

	store := r.backend(m.Type)
	return store.Store(ctx, m)
}

// Update routes updates to the correct backend.
func (r *Router) Update(ctx context.Context, m *Memory) error {
	store := r.backend(m.Type)
	return store.Update(ctx, m)
}

// Delete routes deletes to the correct backend.
func (r *Router) Delete(ctx context.Context, id uuid.UUID, memType MemoryType) error {
	store := r.backend(memType)
	return store.Delete(ctx, id)
}

// Retrieve routes retrieval.
func (r *Router) Retrieve(ctx context.Context, id uuid.UUID, memType MemoryType, opts RetrieveOpts) (*Memory, error) {
	store := r.backend(memType)
	return store.Retrieve(ctx, id, opts)
}

// Search routes search.
func (r *Router) Search(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	hasWorking := false
	hasPersistent := false

	for _, t := range query.Types {
		if t == Working {
			hasWorking = true
		} else {
			hasPersistent = true
		}
	}

	if len(query.Types) == 0 {
		hasWorking = true
		hasPersistent = true
	}

	var results []Memory
	var total int

	if hasPersistent {
		res, err := r.postgres.Search(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("postgres search: %w", err)
		}
		results = append(results, res.Results...)
		total += res.Total
	}

	if hasWorking {
		res, err := r.valkey.Search(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("valkey search: %w", err)
		}
		results = append(results, res.Results...)
		total += res.Total
	}

	// For simplicity in v0.1, just return appended list. In full impl, merge and sort.
	return &SearchResult{
		Results: results,
		Total:   total,
		Limit:   query.Limit,
		Offset:  query.Offset,
	}, nil
}

// ListByNamespace lists memories in a namespace across backends.
func (r *Router) ListByNamespace(ctx context.Context, namespace string, opts ListOpts) (*SearchResult, error) {
	hasWorking := false
	hasPersistent := false

	for _, t := range opts.Types {
		if t == Working {
			hasWorking = true
		} else {
			hasPersistent = true
		}
	}

	if len(opts.Types) == 0 {
		hasWorking = true
		hasPersistent = true
	}

	var results []Memory
	var total int

	if hasPersistent {
		res, err := r.postgres.ListByNamespace(ctx, namespace, opts)
		if err != nil {
			return nil, fmt.Errorf("postgres list: %w", err)
		}
		results = append(results, res.Results...)
		total += res.Total
	}

	if hasWorking {
		res, err := r.valkey.ListByNamespace(ctx, namespace, opts)
		if err != nil {
			return nil, fmt.Errorf("valkey list: %w", err)
		}
		results = append(results, res.Results...)
		total += res.Total
	}

	return &SearchResult{
		Results: results,
		Total:   total,
		Limit:   opts.Limit,
		Offset:  opts.Offset,
	}, nil
}

// RedactPII redacts PII from all backends for a namespace.
func (r *Router) RedactPII(ctx context.Context, namespace string, piiTypes []string) (int, error) {
	pgCount, err := r.postgres.RedactPII(ctx, namespace, piiTypes)
	if err != nil {
		return 0, fmt.Errorf("postgres redact: %w", err)
	}

	vkCount, err := r.valkey.RedactPII(ctx, namespace, piiTypes)
	if err != nil {
		return 0, fmt.Errorf("valkey redact: %w", err)
	}

	return pgCount + vkCount, nil
}

// Expire removes memories past their ValidUntil from all backends.
func (r *Router) Expire(ctx context.Context, before time.Time) (int, error) {
	pgCount, err := r.postgres.Expire(ctx, before)
	if err != nil {
		return 0, fmt.Errorf("postgres expire: %w", err)
	}

	vkCount, err := r.valkey.Expire(ctx, before)
	if err != nil {
		return 0, fmt.Errorf("valkey expire: %w", err)
	}

	return pgCount + vkCount, nil
}

// Ping checks connectivity to all backends.
func (r *Router) Ping(ctx context.Context) error {
	if err := r.postgres.Ping(ctx); err != nil {
		return fmt.Errorf("postgres ping: %w", err)
	}
	if err := r.valkey.Ping(ctx); err != nil {
		return fmt.Errorf("valkey ping: %w", err)
	}
	return nil
}

package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valkey-io/valkey-go"
)

// ValkeyStore implements Store for Working Memory with TTL.
type ValkeyStore struct {
	client valkey.Client
	ttl    time.Duration // default TTL for Working Memory
}

// NewValkeyStore creates a Valkey-backed store.
func NewValkeyStore(client valkey.Client, defaultTTL time.Duration) *ValkeyStore {
	if defaultTTL == 0 {
		defaultTTL = 24 * time.Hour
	}
	return &ValkeyStore{
		client: client,
		ttl:    defaultTTL,
	}
}

// key generates a Valkey key for a memory.
func (s *ValkeyStore) key(namespace, agentDID string, id uuid.UUID) string {
	return fmt.Sprintf("acumius:working:%s:%s:%s", namespace, agentDID, id.String())
}

// latestKey generates a sorted set key for tracking latest memories.
func (s *ValkeyStore) latestKey(namespace, agentDID string) string {
	return fmt.Sprintf("acumius:working:%s:%s:latest", namespace, agentDID)
}

// Store saves Working Memory to Valkey with TTL.
func (s *ValkeyStore) Store(ctx context.Context, m *Memory) error {
	if m.Type != Working {
		return fmt.Errorf("valkey store only supports Working memory, got %s", m.Type)
	}

	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal memory: %w", err)
	}

	key := s.key(m.Namespace, m.AgentDID, m.ID)

	ttlSeconds := int(s.ttl.Seconds())
	if m.ValidUntil != nil {
		ttlSeconds = int(time.Until(*m.ValidUntil).Seconds())
	}

	err = s.client.Do(ctx, s.client.B().Set().Key(key).Value(string(data)).ExSeconds(int64(ttlSeconds)).Build()).Error()
	if err != nil {
		return fmt.Errorf("valkey set: %w", err)
	}

	// Maintain index
	indexKey := fmt.Sprintf("acumius:working:index:%s", m.ID.String())
	err = s.client.Do(ctx, s.client.B().Set().Key(indexKey).Value(key).ExSeconds(int64(ttlSeconds)).Build()).Error()
	if err != nil {
		return fmt.Errorf("valkey set index: %w", err)
	}

	latestKey := s.latestKey(m.Namespace, m.AgentDID)
	err = s.client.Do(ctx, s.client.B().Arbitrary("ZADD", latestKey, fmt.Sprintf("%f", float64(m.CreatedAt.Unix())), m.ID.String()).Build()).Error()
	if err != nil {
		return fmt.Errorf("valkey zadd: %w", err)
	}

	// Best-effort trim of the latest-index set; ignore error.
	_ = s.client.Do(ctx, s.client.B().Zremrangebyrank().Key(latestKey).Start(0).Stop(-1001).Build()).Error()

	return nil
}

// Retrieve fetches Working Memory by ID.
func (s *ValkeyStore) Retrieve(ctx context.Context, id uuid.UUID, opts RetrieveOpts) (*Memory, error) {
	indexKey := fmt.Sprintf("acumius:working:index:%s", id.String())

	actualKey, err := s.client.Do(ctx, s.client.B().Get().Key(indexKey).Build()).ToString()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return nil, fmt.Errorf("memory not found: %s", id)
		}
		return nil, fmt.Errorf("valkey get index: %w", err)
	}

	data, err := s.client.Do(ctx, s.client.B().Get().Key(actualKey).Build()).ToString()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return nil, fmt.Errorf("memory expired or deleted: %s", id)
		}
		return nil, fmt.Errorf("valkey get: %w", err)
	}

	var m Memory
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		return nil, fmt.Errorf("unmarshal memory: %w", err)
	}

	return &m, nil
}

// Search performs keyword search on Working Memory.
func (s *ValkeyStore) Search(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	pattern := "acumius:working:*:*:*"
	if len(query.Namespaces) == 1 {
		pattern = fmt.Sprintf("acumius:working:%s:*:*", query.Namespaces[0])
	}

	var results []Memory
	var cursor uint64

	for {
		scanResult, err := s.client.Do(ctx, s.client.B().Scan().Cursor(cursor).Match(pattern).Count(100).Build()).AsScanEntry()
		if err != nil {
			return nil, fmt.Errorf("valkey scan: %w", err)
		}

		for _, key := range scanResult.Elements {
			// skip index keys
			if strings.Contains(key, ":index:") || strings.HasSuffix(key, ":latest") {
				continue
			}

			data, err := s.client.Do(ctx, s.client.B().Get().Key(key).Build()).ToString()
			if err != nil {
				continue
			}

			var m Memory
			if err := json.Unmarshal([]byte(data), &m); err != nil {
				continue
			}

			if !strings.Contains(string(m.Content), query.Query) {
				continue
			}

			if len(query.Namespaces) > 0 && !containsString(query.Namespaces, m.Namespace) {
				continue
			}

			results = append(results, m)

			if len(results) >= query.Limit {
				break
			}
		}

		cursor = scanResult.Cursor
		if cursor == 0 || len(results) >= query.Limit {
			break
		}
	}

	return &SearchResult{
		Results: results,
		Total:   len(results),
		Limit:   query.Limit,
		Offset:  query.Offset,
	}, nil
}

// ListByNamespace lists Working Memory in a namespace.
func (s *ValkeyStore) ListByNamespace(ctx context.Context, namespace string, opts ListOpts) (*SearchResult, error) {
	// Simplified scan
	return &SearchResult{}, nil
}

// Delete removes Working Memory.
func (s *ValkeyStore) Delete(ctx context.Context, id uuid.UUID) error {
	indexKey := fmt.Sprintf("acumius:working:index:%s", id.String())

	actualKey, err := s.client.Do(ctx, s.client.B().Get().Key(indexKey).Build()).ToString()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return fmt.Errorf("memory not found: %s", id)
		}
		return fmt.Errorf("valkey get index: %w", err)
	}

	err = s.client.Do(ctx, s.client.B().Del().Key(indexKey).Build()).Error()
	if err != nil {
		return fmt.Errorf("valkey del index: %w", err)
	}

	err = s.client.Do(ctx, s.client.B().Del().Key(actualKey).Build()).Error()
	if err != nil {
		return fmt.Errorf("valkey del: %w", err)
	}

	return nil
}

func (s *ValkeyStore) Update(ctx context.Context, m *Memory) error {
	return s.Store(ctx, m)
}

func (s *ValkeyStore) RedactPII(ctx context.Context, namespace string, piiTypes []string) (int, error) {
	return 0, nil
}

func (s *ValkeyStore) Expire(ctx context.Context, before time.Time) (int, error) {
	return 0, nil
}

func (s *ValkeyStore) Ping(ctx context.Context) error {
	return s.client.Do(ctx, s.client.B().Ping().Build()).Error()
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

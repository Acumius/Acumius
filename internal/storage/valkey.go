package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/valkey-io/valkey-go"
)

type ValkeyStore struct {
	client valkey.Client
}

func NewValkeyStore(connectionString string) (*ValkeyStore, error) {
	addr := connectionString
	if strings.HasPrefix(addr, "valkey://") {
		addr = strings.TrimPrefix(addr, "valkey://")
	}

	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{addr}})
	if err != nil {
		return nil, fmt.Errorf("failed to create valkey client: %w", err)
	}

	return &ValkeyStore{client: client}, nil
}

func (s *ValkeyStore) Ping(ctx context.Context) error {
	cmd := s.client.B().Ping().Build()
	err := s.client.Do(ctx, cmd).Error()
	if err != nil {
		return fmt.Errorf("valkey ping failed: %w", err)
	}
	return nil
}

func (s *ValkeyStore) Close() {
	s.client.Close()
}

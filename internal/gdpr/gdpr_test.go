package gdpr_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/Acumius/Acumius/internal/gdpr"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestGDPRService(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Clean tables
	_, _ = db.Exec("TRUNCATE agents, memories, policies CASCADE")

	agentDID := "did:example:gdpr-123"

	// Seed agent
	_, err := db.ExecContext(ctx, "INSERT INTO agents (did, public_key, reputation_score, created_at, last_active_at) VALUES ($1, $2, 500, NOW(), NOW())", agentDID, "pubkey")
	require.NoError(t, err)

	// Seed memory
	memoryID := "123e4567-e89b-12d3-a456-426614174000"
	_, err = db.ExecContext(ctx, "INSERT INTO memories (id, agent_did, type, namespace, content, created_at) VALUES ($1, $2, 'episodic', 'default', '{\"text\":\"hello\"}', NOW())", memoryID, agentDID)
	require.NoError(t, err)

	// Seed policy
	_, err = db.ExecContext(ctx, "INSERT INTO policies (id, agent_did, content, version, created_at, updated_at) VALUES ($1, $2, '{}', '1.0', NOW(), NOW())", "pol-1", agentDID)
	require.NoError(t, err)

	svc := gdpr.NewService(db)

	// 1. Export Data
	exported, err := svc.Export(ctx, agentDID)
	require.NoError(t, err)
	assert.Contains(t, string(exported), memoryID)
	assert.Contains(t, string(exported), "hello")

	// 2. Forget (soft delete memory, delete policy, revoke agent)
	err = svc.Forget(ctx, agentDID)
	require.NoError(t, err)

	// Verify agent revoked
	var pubKey []byte
	var repScore int
	err = db.QueryRow("SELECT public_key, reputation_score FROM agents WHERE did = $1", agentDID).Scan(&pubKey, &repScore)
	require.NoError(t, err)
	assert.Equal(t, 0, repScore)
	assert.Empty(t, pubKey)

	// Verify policy deleted
	var polCount int
	err = db.QueryRow("SELECT COUNT(*) FROM policies WHERE agent_did = $1", agentDID).Scan(&polCount)
	require.NoError(t, err)
	assert.Equal(t, 0, polCount)

	// Verify memory soft deleted
	var deletedAt sql.NullTime
	err = db.QueryRow("SELECT deleted_at FROM memories WHERE id = $1", memoryID).Scan(&deletedAt)
	require.NoError(t, err)
	assert.True(t, deletedAt.Valid)
	assert.True(t, deletedAt.Time.Before(time.Now().Add(time.Second)))

	// 3. Export Data after Forget (should be empty as it checks deleted_at IS NULL)
	exported, err = svc.Export(ctx, agentDID)
	require.NoError(t, err)
	assert.Equal(t, "null", string(exported)) // json.Marshal of empty slice is null if nil, or [] if empty. Here it's nil.
}

//go:build integration

package storage

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPoolAndMigrate(t *testing.T) {
	ctx := context.Background()
	url := os.Getenv("SQT_DATABASE_URL")
	if url == "" {
		t.Skip("SQT_DATABASE_URL not set")
	}

	pool, err := NewPool(ctx, url)
	require.NoError(t, err)
	defer pool.Close()

	// Clean slate for the test.
	require.NoError(t, MigrateDown(pool))

	require.NoError(t, MigrateUp(pool))

	var tableExists bool
	err = pool.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'audit_records')",
	).Scan(&tableExists)
	require.NoError(t, err)
	assert.True(t, tableExists, "audit_records table should exist after migration")

	var indexCount int
	err = pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public' AND tablename = 'audit_records'",
	).Scan(&indexCount)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, indexCount, 2, "audit_records should have at least the two expected indexes")

	// Teardown.
	require.NoError(t, MigrateDown(pool))
}

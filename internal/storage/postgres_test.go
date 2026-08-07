//go:build integration

package storage

import (
	"context"
	"os"
	"testing"

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

	err = MigrateUp(pool)
	require.NoError(t, err)
}

//go:build integration

package audit

import (
	"context"
	"os"
	"testing"

	"github.com/kevincouton/standard-tools-go/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresStorageRoundTrip(t *testing.T) {
	ctx := context.Background()
	url := os.Getenv("SQT_DATABASE_URL")
	if url == "" {
		t.Skip("SQT_DATABASE_URL not set")
	}

	pool, err := storage.NewPool(ctx, url)
	require.NoError(t, err)
	defer pool.Close()

	require.NoError(t, storage.MigrateDown(pool))
	require.NoError(t, storage.MigrateUp(pool))

	store := NewPostgresStorage(pool)
	w := NewWriter(store)
	v := NewVerifier(store)

	require.NoError(t, w.Write(ctx, sampleRecord("r1")))
	require.NoError(t, w.Write(ctx, sampleRecord("r2")))

	assert.NoError(t, v.VerifyChain(ctx))

	latest, err := store.Latest(ctx)
	require.NoError(t, err)
	assert.Equal(t, "r2", latest.RequestID)

	byID, err := store.GetByRequestID(ctx, "r1")
	require.NoError(t, err)
	assert.Equal(t, "r1", byID.RequestID)

	require.NoError(t, storage.MigrateDown(pool))
}

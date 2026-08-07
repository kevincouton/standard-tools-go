package audit

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleRecord(id string) DecisionRecord {
	return DecisionRecord{
		RequestID:      id,
		RecordedAt:     time.Date(2026, 8, 7, 0, 0, 0, 0, time.UTC),
		ToolName:       "list_tools",
		Input:          map[string]any{"ticker": "AAPL"},
		Output:         map[string]any{"tools": []string{"list_tools"}},
		Status:         "ok",
		GitCommitSHA:   "abc123",
		PackageVersion: "0.1.0",
		RandomSeed:     42,
	}
}

func TestHashRecord_Stability(t *testing.T) {
	r := sampleRecord("r1")
	h1, err := HashRecord(r)
	require.NoError(t, err)
	h2, err := HashRecord(r)
	require.NoError(t, err)
	assert.Equal(t, h1, h2, "hash should be stable across identical records")

	r.RecordHash = "tampered"
	h3, err := HashRecord(r)
	require.NoError(t, err)
	assert.Equal(t, h1, h3, "HashRecord should ignore existing RecordHash value")
}

func TestWriter_ChainsRecords(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()
	w := NewWriter(store)

	r1 := sampleRecord("r1")
	require.NoError(t, w.Write(ctx, r1))

	stored1, err := store.Latest(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, stored1.RecordHash, "first record should have a record hash")
	assert.Empty(t, stored1.PrevRecordHash, "first record should have no previous hash")

	r2 := sampleRecord("r2")
	require.NoError(t, w.Write(ctx, r2))

	stored2, err := store.Latest(ctx)
	require.NoError(t, err)
	assert.Equal(t, stored1.RecordHash, stored2.PrevRecordHash, "second record should chain to first")
	assert.NotEqual(t, stored1.RecordHash, stored2.RecordHash, "record hashes should differ")
}

func TestWriter_NormalizesRecordedAtToUTC(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()
	w := NewWriter(store)

	r := sampleRecord("r1")
	r.RecordedAt = time.Date(2026, 8, 7, 12, 0, 0, 0, time.FixedZone("IST", 5*60*60))
	require.NoError(t, w.Write(ctx, r))

	stored, err := store.Latest(ctx)
	require.NoError(t, err)
	assert.Equal(t, time.UTC, stored.RecordedAt.Location())
}

func TestWriter_ComputesInputOutputHashes(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()
	w := NewWriter(store)

	r := sampleRecord("r1")
	require.NoError(t, w.Write(ctx, r))

	stored, err := store.Latest(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, stored.InputHash)
	assert.NotEmpty(t, stored.OutputHash)
}

func TestVerifier_ValidChain(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()
	w := NewWriter(store)
	v := NewVerifier(store)

	require.NoError(t, w.Write(ctx, sampleRecord("r1")))
	require.NoError(t, w.Write(ctx, sampleRecord("r2")))

	assert.NoError(t, v.VerifyChain(ctx))
}

func TestVerifier_EmptyChain(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()
	v := NewVerifier(store)

	assert.NoError(t, v.VerifyChain(ctx))
}

func TestVerifier_TamperedRecord(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()
	w := NewWriter(store)
	v := NewVerifier(store)

	require.NoError(t, w.Write(ctx, sampleRecord("r1")))

	latest, err := store.Latest(ctx)
	require.NoError(t, err)
	latest.Status = "malicious"
	require.NoError(t, store.Append(ctx, latest))

	assert.Error(t, v.VerifyChain(ctx))
}

func TestVerifier_TamperedInputHash(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()
	w := NewWriter(store)
	v := NewVerifier(store)

	require.NoError(t, w.Write(ctx, sampleRecord("r1")))

	latest, err := store.Latest(ctx)
	require.NoError(t, err)
	latest.InputHash = "tampered"
	require.NoError(t, store.Append(ctx, latest))

	assert.Error(t, v.VerifyChain(ctx))
}

func TestMemoryStorage_GetByRequestID(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()

	r1 := sampleRecord("r1")
	r2 := sampleRecord("r2")
	require.NoError(t, store.Append(ctx, r1))
	require.NoError(t, store.Append(ctx, r2))

	found, err := store.GetByRequestID(ctx, "r1")
	require.NoError(t, err)
	assert.Equal(t, "r1", found.RequestID)

	_, err = store.GetByRequestID(ctx, "unknown")
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestMemoryStorage_LatestNotFound(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()

	_, err := store.Latest(ctx)
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestWriter_ConcurrentWrites(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()
	w := NewWriter(store)
	v := NewVerifier(store)

	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			r := sampleRecord(fmt.Sprintf("r%d", i))
			_ = w.Write(ctx, r)
		}(i)
	}
	wg.Wait()

	require.NoError(t, v.VerifyChain(ctx))

	latest, err := store.Latest(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, latest.RecordHash)

	// Count distinct prev_record_hash values: each record (except the first) should
	// point to exactly one predecessor, and no two records should share the same prev.
	records := make(map[string]DecisionRecord)
	for i := 0; i < n; i++ {
		r, err := store.GetByRequestID(ctx, fmt.Sprintf("r%d", i))
		require.NoError(t, err)
		records[r.RequestID] = r
	}

	prevs := make(map[string]int)
	for _, r := range records {
		if r.PrevRecordHash != "" {
			prevs[r.PrevRecordHash]++
		}
	}
	for hash, count := range prevs {
		assert.Equal(t, 1, count, "prev_record_hash %s is shared by multiple records", hash)
	}
}

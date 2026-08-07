package audit

import (
	"context"
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

func TestVerifier_ValidChain(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()
	w := NewWriter(store)
	v := NewVerifier(store)

	require.NoError(t, w.Write(ctx, sampleRecord("r1")))
	require.NoError(t, w.Write(ctx, sampleRecord("r2")))

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

	missing, err := store.GetByRequestID(ctx, "unknown")
	require.NoError(t, err)
	assert.Equal(t, DecisionRecord{}, missing)
}

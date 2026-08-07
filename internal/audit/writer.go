package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Writer records hash-chained decision records to the underlying storage.
type Writer struct {
	storage Storage
}

func NewWriter(storage Storage) *Writer {
	return &Writer{storage: storage}
}

// Write computes hashes and chains the record to the previous one before persisting it.
// Input and Output are canonicalized to json.RawMessage so that hashing and later
// verification are stable across storage round-trips.
func (w *Writer) Write(ctx context.Context, r DecisionRecord) error {
	if r.RecordedAt.IsZero() {
		r.RecordedAt = Now()
	}
	r.RecordedAt = r.RecordedAt.UTC()

	latest, err := w.storage.Latest(ctx)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("fetch latest audit record: %w", err)
	}
	if err == nil && latest.RecordHash != "" {
		r.PrevRecordHash = latest.RecordHash
	}

	inputJSON, inputHash, err := canonicalizeAndHash(r.Input)
	if err != nil {
		return fmt.Errorf("hash input: %w", err)
	}
	r.Input = inputJSON
	r.InputHash = inputHash

	outputJSON, outputHash, err := canonicalizeAndHash(r.Output)
	if err != nil {
		return fmt.Errorf("hash output: %w", err)
	}
	r.Output = outputJSON
	r.OutputHash = outputHash

	recordHash, err := HashRecord(r)
	if err != nil {
		return fmt.Errorf("hash record: %w", err)
	}
	r.RecordHash = recordHash

	if err := w.storage.Append(ctx, r); err != nil {
		return fmt.Errorf("append audit record: %w", err)
	}
	return nil
}

// canonicalizeAndHash marshals v to JSON and returns both the canonical json.RawMessage
// and its SHA-256 hash. Using json.RawMessage for storage ensures that re-marshaling
// later produces byte-identical output, keeping hashes stable.
func canonicalizeAndHash(v any) (json.RawMessage, string, error) {
	if v == nil {
		return json.RawMessage("null"), hashBytes([]byte("null")), nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, "", fmt.Errorf("marshal: %w", err)
	}
	return json.RawMessage(b), hashBytes(b), nil
}

// Now is a testable clock for RecordedAt. Tests may override it.
var Now = func() time.Time {
	return time.Now().UTC()
}

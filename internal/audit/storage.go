package audit

import (
	"context"
	"sync"
)

// Storage persists decision records. Implementations must be safe for concurrent use.
type Storage interface {
	Append(ctx context.Context, r DecisionRecord) error
	Latest(ctx context.Context) (DecisionRecord, error)
	GetByRequestID(ctx context.Context, requestID string) (DecisionRecord, error)
}

// MemoryStorage is an in-memory implementation for unit tests.
type MemoryStorage struct {
	mu      sync.Mutex
	records []DecisionRecord
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

func (m *MemoryStorage) Append(_ context.Context, r DecisionRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records = append(m.records, r)
	return nil
}

func (m *MemoryStorage) Latest(_ context.Context) (DecisionRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.records) == 0 {
		return DecisionRecord{}, nil
	}
	return m.records[len(m.records)-1], nil
}

func (m *MemoryStorage) GetByRequestID(_ context.Context, requestID string) (DecisionRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := len(m.records) - 1; i >= 0; i-- {
		if m.records[i].RequestID == requestID {
			return m.records[i], nil
		}
	}
	return DecisionRecord{}, nil
}

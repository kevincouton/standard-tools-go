package marketdata

import (
	"context"
	"sync"

	"github.com/kevincouton/standard-tools-go/internal/core"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]core.OHLCV, bool)
	Put(ctx context.Context, key string, series []core.OHLCV)
}

type InMemoryCache struct {
	mu   sync.RWMutex
	data map[string][]core.OHLCV
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{data: make(map[string][]core.OHLCV)}
}

func (c *InMemoryCache) Get(ctx context.Context, key string) ([]core.OHLCV, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	series, ok := c.data[key]
	return series, ok
}

func (c *InMemoryCache) Put(ctx context.Context, key string, series []core.OHLCV) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = series
}

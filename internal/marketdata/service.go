package marketdata

import (
	"context"
	"fmt"
	"sync"

	"github.com/kevincouton/standard-tools-go/internal/core"
)

type Service struct {
	defaultProvider string
	providers       map[string]Provider
	mu              sync.RWMutex
	cache           Cache
}

func NewService(defaultProvider string, cache Cache) *Service {
	return &Service{
		defaultProvider: defaultProvider,
		providers:       make(map[string]Provider),
		cache:           cache,
	}
}

func (s *Service) Register(provider Provider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[provider.Name()] = provider
}

func (s *Service) Fetch(ctx context.Context, ticker core.Ticker, interval core.BarInterval, rng core.DateRange, providerName string) ([]core.OHLCV, error) {
	name := s.defaultProvider
	if providerName != "" {
		name = providerName
	}
	s.mu.RLock()
	provider, ok := s.providers[name]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: provider %s not registered", core.ErrProviderNotAvailable, name)
	}
	key := fmt.Sprintf("%s:%s:%s:%s:%s", name, ticker.Symbol, interval.String(), rng.Start.Format("2006-01-02"), rng.End.Format("2006-01-02"))
	if cached, hit := s.cache.Get(ctx, key); hit {
		return cached, nil
	}
	series, err := provider.Fetch(ctx, ticker, interval, rng)
	if err != nil {
		return nil, err
	}
	s.cache.Put(ctx, key, series)
	return series, nil
}

package marketdata

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kevincouton/standard-tools-go/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestServiceFetchesSyntheticData(t *testing.T) {
	svc := NewService("synthetic", NewInMemoryCache())
	svc.Register(&SyntheticProvider{})
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
	rng, _ := core.NewDateRange(start, end)
	ticker, _ := core.NewTicker("TEST")
	series, err := svc.Fetch(context.Background(), ticker, core.Daily, rng, "")
	assert.NoError(t, err)
	assert.Len(t, series, 5)
}

func TestServiceCachesResults(t *testing.T) {
	cache := NewInMemoryCache()
	svc := NewService("synthetic", cache)
	svc.Register(&SyntheticProvider{})
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
	rng, _ := core.NewDateRange(start, end)
	ticker, _ := core.NewTicker("TEST")

	first, err := svc.Fetch(context.Background(), ticker, core.Daily, rng, "")
	assert.NoError(t, err)
	first[0].Volume = 999_999

	second, err := svc.Fetch(context.Background(), ticker, core.Daily, rng, "")
	assert.NoError(t, err)
	assert.Equal(t, int64(1_000_000), second[0].Volume)
}

func TestServiceProviderNotRegistered(t *testing.T) {
	svc := NewService("missing", NewInMemoryCache())
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
	rng, _ := core.NewDateRange(start, end)
	ticker, _ := core.NewTicker("TEST")
	_, err := svc.Fetch(context.Background(), ticker, core.Daily, rng, "")
	assert.True(t, errors.Is(err, core.ErrProviderNotAvailable))
}

func TestServiceExplicitProviderOverride(t *testing.T) {
	svc := NewService("yahoo", NewInMemoryCache())
	svc.Register(&SyntheticProvider{})
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
	rng, _ := core.NewDateRange(start, end)
	ticker, _ := core.NewTicker("TEST")
	series, err := svc.Fetch(context.Background(), ticker, core.Daily, rng, "synthetic")
	assert.NoError(t, err)
	assert.Len(t, series, 5)
}

func TestYahooProviderNotImplemented(t *testing.T) {
	y := &YahooProvider{}
	assert.Equal(t, "yahoo", y.Name())
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
	rng, _ := core.NewDateRange(start, end)
	ticker, _ := core.NewTicker("TEST")
	_, err := y.Fetch(context.Background(), ticker, core.Daily, rng)
	assert.True(t, errors.Is(err, core.ErrProviderNotAvailable))
}

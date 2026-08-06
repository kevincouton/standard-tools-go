package marketdata

import (
	"context"
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
	series, err := svc.Fetch(context.Background(), ticker, core.Daily, rng, nil)
	assert.NoError(t, err)
	assert.Len(t, series, 5)
}

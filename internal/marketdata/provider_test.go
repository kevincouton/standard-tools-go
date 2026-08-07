package marketdata

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kevincouton/standard-tools-go/internal/core"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewProviderFactory(t *testing.T) {
	cases := []struct {
		name    string
		wantErr bool
	}{
		{"synthetic", false},
		{"yahoo", false},
		{"polygon", false},
		{"unknown", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProvider(tc.name)
			if tc.wantErr {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, core.ErrProviderNotAvailable))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.name, p.Name())
		})
	}
}

func TestYahooProviderFetchWithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v8/finance/chart/TEST", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"chart": {
				"result": [{
					"timestamp": [1704067200, 1704153600],
					"indicators": {
						"quote": [{
							"open": [100.0, 101.0],
							"high": [102.0, 103.0],
							"low": [99.0, 100.0],
							"close": [101.0, 102.0],
							"volume": [1000, 2000]
						}]
					}
				}],
				"error": null
			}
		}`))
	}))
	defer server.Close()

	provider := NewYahooProvider()
	provider.client = server.Client()
	provider.baseURL = server.URL + "/v8/finance/chart"

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	rng, _ := core.NewDateRange(start, end)
	ticker, _ := core.NewTicker("TEST")

	bars, err := provider.Fetch(context.Background(), ticker, core.Daily, rng)
	assert.NoError(t, err)
	assert.Len(t, bars, 2)
}

func TestYahooProviderConvertResult(t *testing.T) {
	provider := NewYahooProvider()
	result := yahooChartResult{
		Timestamp: []int64{1704067200, 1704153600},
		Indicators: struct {
			Quote []struct {
				Open   []float64 `json:"open"`
				High   []float64 `json:"high"`
				Low    []float64 `json:"low"`
				Close  []float64 `json:"close"`
				Volume []int64   `json:"volume"`
			} `json:"quote"`
		}{
			Quote: []struct {
				Open   []float64 `json:"open"`
				High   []float64 `json:"high"`
				Low    []float64 `json:"low"`
				Close  []float64 `json:"close"`
				Volume []int64   `json:"volume"`
			}{
				{
					Open:   []float64{100.0, 101.0},
					High:   []float64{102.0, 103.0},
					Low:    []float64{99.0, 100.0},
					Close:  []float64{101.0, 102.0},
					Volume: []int64{1000, 2000},
				},
			},
		},
	}

	bars, err := provider.convertResult(result)
	assert.NoError(t, err)
	assert.Len(t, bars, 2)
	assert.True(t, bars[0].Open.Equal(decimal.NewFromInt(100)))
	assert.True(t, bars[0].High.Equal(decimal.NewFromInt(102)))
	assert.True(t, bars[0].Low.Equal(decimal.NewFromInt(99)))
	assert.True(t, bars[0].Close.Equal(decimal.NewFromInt(101)))
	assert.Equal(t, int64(1000), bars[0].Volume)
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), bars[0].Date)
}

func TestPolygonProviderFetchWithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/aggs/ticker/TEST/range/1/day/2024-01-01/2024-01-02", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, "true", q.Get("adjusted"))
		assert.Equal(t, "asc", q.Get("sort"))
		assert.Equal(t, "test-api-key", q.Get("apiKey"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"results": [
				{"t": 1704067200000, "o": 100.0, "h": 102.0, "l": 99.0, "c": 101.0, "v": 1000},
				{"t": 1704153600000, "o": 101.0, "h": 103.0, "l": 100.0, "c": 102.0, "v": 2000}
			]
		}`))
	}))
	defer server.Close()

	provider := NewPolygonProvider("test-api-key")
	provider.client = server.Client()
	provider.baseURL = server.URL

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	rng, _ := core.NewDateRange(start, end)
	ticker, _ := core.NewTicker("TEST")

	bars, err := provider.Fetch(context.Background(), ticker, core.Daily, rng)
	assert.NoError(t, err)
	assert.Len(t, bars, 2)
}

func TestPolygonProviderConvertResults(t *testing.T) {
	provider := NewPolygonProvider("test-api-key")
	results := []polygonAggregate{
		{Timestamp: 1704067200000, Open: 100.0, High: 102.0, Low: 99.0, Close: 101.0, Volume: 1000},
		{Timestamp: 1704153600000, Open: 101.0, High: 103.0, Low: 100.0, Close: 102.0, Volume: 2000},
	}

	bars, err := provider.convertResults(results)
	assert.NoError(t, err)
	assert.Len(t, bars, 2)
	assert.True(t, bars[0].Open.Equal(decimal.NewFromInt(100)))
	assert.True(t, bars[0].High.Equal(decimal.NewFromInt(102)))
	assert.True(t, bars[0].Low.Equal(decimal.NewFromInt(99)))
	assert.True(t, bars[0].Close.Equal(decimal.NewFromInt(101)))
	assert.Equal(t, int64(1000), bars[0].Volume)
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), bars[0].Date)
}

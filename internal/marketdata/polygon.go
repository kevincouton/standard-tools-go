package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/kevincouton/standard-tools-go/internal/core"
	"github.com/shopspring/decimal"
)

const defaultPolygonBaseURL = "https://api.polygon.io"

type PolygonProvider struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewPolygonProvider(apiKey string) *PolygonProvider {
	return &PolygonProvider{
		baseURL: defaultPolygonBaseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (p *PolygonProvider) Name() string { return "polygon" }

func (p *PolygonProvider) Fetch(ctx context.Context, ticker core.Ticker, interval core.BarInterval, rng core.DateRange) ([]core.OHLCV, error) {
	multiplier, timespan := p.mapInterval(interval)
	from := rng.Start.Format("2006-01-02")
	to := rng.End.Format("2006-01-02")

	u, err := url.Parse(fmt.Sprintf("%s/v2/aggs/ticker/%s/range/%d/%s/%s/%s",
		p.baseURL, url.PathEscape(ticker.Symbol), multiplier, timespan, from, to))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid polygon url: %w", core.ErrProviderNotAvailable, err)
	}
	q := u.Query()
	q.Set("adjusted", "true")
	q.Set("sort", "asc")
	q.Set("apiKey", p.apiKey)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to build polygon request: %w", core.ErrProviderNotAvailable, err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: polygon request failed: %w", core.ErrProviderNotAvailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: polygon returned status %d", core.ErrProviderNotAvailable, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read polygon response: %w", core.ErrProviderNotAvailable, err)
	}

	var payload polygonAggregatesResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("%w: failed to parse polygon response: %w", core.ErrProviderNotAvailable, err)
	}

	return p.convertResults(payload.Results)
}

func (p *PolygonProvider) GetTickerInfo(ctx context.Context, ticker core.Ticker) (core.TickerInfo, error) {
	return core.TickerInfo{}, fmt.Errorf("%w: polygon ticker info not available", core.ErrProviderNotAvailable)
}

func (p *PolygonProvider) GetFinancialRatios(ctx context.Context, ticker core.Ticker) (core.FinancialRatios, error) {
	return core.FinancialRatios{}, fmt.Errorf("%w: polygon financial ratios not available", core.ErrProviderNotAvailable)
}

func (p *PolygonProvider) GetMetadata(ctx context.Context) (core.DataSetMetadata, error) {
	return core.DataSetMetadata{
		Provider:         "polygon",
		Adjusted:         true,
		SurvivorshipFree: true,
		PointInTime:      false,
		Frequency:        "daily",
		Timezone:         "America/New_York",
		RetrievedAt:      time.Now().UTC(),
	}, nil
}

func (p *PolygonProvider) mapInterval(interval core.BarInterval) (int, string) {
	switch interval {
	case core.Weekly:
		return 1, "week"
	case core.Monthly:
		return 1, "month"
	default:
		return 1, "day"
	}
}

func (p *PolygonProvider) convertResults(results []polygonAggregate) ([]core.OHLCV, error) {
	bars := make([]core.OHLCV, 0, len(results))
	for _, r := range results {
		bars = append(bars, core.OHLCV{
			Date:   time.UnixMilli(r.Timestamp).UTC(),
			Open:   decimal.NewFromFloat(r.Open),
			High:   decimal.NewFromFloat(r.High),
			Low:    decimal.NewFromFloat(r.Low),
			Close:  decimal.NewFromFloat(r.Close),
			Volume: int64(r.Volume),
		})
	}
	return bars, nil
}

type polygonAggregatesResponse struct {
	Results []polygonAggregate `json:"results"`
}

type polygonAggregate struct {
	Timestamp int64   `json:"t"`
	Open      float64 `json:"o"`
	High      float64 `json:"h"`
	Low       float64 `json:"l"`
	Close     float64 `json:"c"`
	Volume    float64 `json:"v"`
}

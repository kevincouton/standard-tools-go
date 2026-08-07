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

const defaultYahooChartURL = "https://query1.finance.yahoo.com/v8/finance/chart"

type YahooProvider struct {
	baseURL string
	client  *http.Client
}

func NewYahooProvider() *YahooProvider {
	return &YahooProvider{
		baseURL: defaultYahooChartURL,
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (y *YahooProvider) Name() string { return "yahoo" }

func (y *YahooProvider) Fetch(ctx context.Context, ticker core.Ticker, interval core.BarInterval, rng core.DateRange) ([]core.OHLCV, error) {
	intervalStr := y.mapInterval(interval)
	u, err := url.Parse(fmt.Sprintf("%s/%s", y.baseURL, url.PathEscape(ticker.Symbol)))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid yahoo url: %w", core.ErrProviderNotAvailable, err)
	}
	q := u.Query()
	q.Set("period1", fmt.Sprintf("%d", rng.Start.Unix()))
	q.Set("period2", fmt.Sprintf("%d", rng.End.Unix()))
	q.Set("interval", intervalStr)
	q.Set("events", "history")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to build yahoo request: %w", core.ErrProviderNotAvailable, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; StandardTools/1.0)")

	resp, err := y.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: yahoo request failed: %w", core.ErrProviderNotAvailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: yahoo returned status %d", core.ErrProviderNotAvailable, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read yahoo response: %w", core.ErrProviderNotAvailable, err)
	}

	var payload yahooChartResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("%w: failed to parse yahoo response: %w", core.ErrProviderNotAvailable, err)
	}

	if len(payload.Chart.Error) > 0 {
		return nil, fmt.Errorf("%w: yahoo api error: %s", core.ErrProviderNotAvailable, payload.Chart.Error)
	}
	if len(payload.Chart.Result) == 0 {
		return nil, fmt.Errorf("%w: yahoo returned empty result", core.ErrProviderNotAvailable)
	}

	result := payload.Chart.Result[0]
	return y.convertResult(result)
}

func (y *YahooProvider) GetTickerInfo(ctx context.Context, ticker core.Ticker) (core.TickerInfo, error) {
	return core.TickerInfo{}, fmt.Errorf("%w: yahoo ticker info not available", core.ErrProviderNotAvailable)
}

func (y *YahooProvider) GetFinancialRatios(ctx context.Context, ticker core.Ticker) (core.FinancialRatios, error) {
	return core.FinancialRatios{}, fmt.Errorf("%w: yahoo financial ratios not available", core.ErrProviderNotAvailable)
}

func (y *YahooProvider) GetMetadata(ctx context.Context) (core.DataSetMetadata, error) {
	return core.DataSetMetadata{
		Provider:         "yahoo",
		Adjusted:         true,
		SurvivorshipFree: false,
		PointInTime:      false,
		Frequency:        "daily",
		Timezone:         "America/New_York",
		RetrievedAt:      time.Now().UTC(),
	}, nil
}

func (y *YahooProvider) mapInterval(interval core.BarInterval) string {
	switch interval {
	case core.Weekly:
		return "1wk"
	case core.Monthly:
		return "1mo"
	default:
		return "1d"
	}
}

func (y *YahooProvider) convertResult(result yahooChartResult) ([]core.OHLCV, error) {
	timestamps := result.Timestamp
	opens := result.Indicators.Quote[0].Open
	highs := result.Indicators.Quote[0].High
	lows := result.Indicators.Quote[0].Low
	closes := result.Indicators.Quote[0].Close
	volumes := result.Indicators.Quote[0].Volume

	n := len(timestamps)
	if n == 0 || n != len(opens) || n != len(highs) || n != len(lows) || n != len(closes) || n != len(volumes) {
		return nil, fmt.Errorf("%w: yahoo result arrays have mismatched lengths", core.ErrProviderNotAvailable)
	}

	bars := make([]core.OHLCV, 0, n)
	for i := 0; i < n; i++ {
		if timestamps[i] == 0 {
			continue
		}
		bars = append(bars, core.OHLCV{
			Date:   time.Unix(timestamps[i], 0).UTC(),
			Open:   decimal.NewFromFloat(opens[i]),
			High:   decimal.NewFromFloat(highs[i]),
			Low:    decimal.NewFromFloat(lows[i]),
			Close:  decimal.NewFromFloat(closes[i]),
			Volume: volumes[i],
		})
	}
	return bars, nil
}

type yahooChartResponse struct {
	Chart struct {
		Result []yahooChartResult `json:"result"`
		Error  string             `json:"error"`
	} `json:"chart"`
}

type yahooChartResult struct {
	Timestamp  []int64 `json:"timestamp"`
	Indicators struct {
		Quote []struct {
			Open   []float64 `json:"open"`
			High   []float64 `json:"high"`
			Low    []float64 `json:"low"`
			Close  []float64 `json:"close"`
			Volume []int64   `json:"volume"`
		} `json:"quote"`
	} `json:"indicators"`
}

package marketdata

import (
	"context"
	"time"

	"github.com/kevincouton/standard-tools-go/internal/core"
	"github.com/shopspring/decimal"
)

type SyntheticProvider struct{}

func (s *SyntheticProvider) Name() string { return "synthetic" }

func (s *SyntheticProvider) Fetch(ctx context.Context, ticker core.Ticker, interval core.BarInterval, rng core.DateRange) ([]core.OHLCV, error) {
	var bars []core.OHLCV
	price := decimal.NewFromInt(100)
	for d := rng.Start; !d.After(rng.End); d = d.AddDate(0, 0, 1) {
		open := price
		close := price.Add(decimal.NewFromInt(1))
		high := decimal.Max(open, close).Add(decimal.NewFromFloat(0.5))
		low := decimal.Min(open, close).Sub(decimal.NewFromFloat(0.5))
		bars = append(bars, core.OHLCV{
			Date:   d,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: 1_000_000,
		})
		price = close
	}
	return bars, nil
}

func (s *SyntheticProvider) GetTickerInfo(ctx context.Context, ticker core.Ticker) (core.TickerInfo, error) {
	return core.TickerInfo{
		Symbol:    ticker.Symbol,
		Name:      ticker.Symbol + " Inc.",
		Sector:    "Technology",
		Industry:  "Software",
		Employees: 1000,
		City:      "New York",
		Country:   "USA",
		Website:   "https://example.com",
	}, nil
}

func (s *SyntheticProvider) GetFinancialRatios(ctx context.Context, ticker core.Ticker) (core.FinancialRatios, error) {
	return core.FinancialRatios{Symbol: ticker.Symbol}, nil
}

func (s *SyntheticProvider) GetMetadata(ctx context.Context) (core.DataSetMetadata, error) {
	return core.DataSetMetadata{
		Provider:         "synthetic",
		Adjusted:         false,
		SurvivorshipFree: false,
		PointInTime:      false,
		Frequency:        "daily",
		Timezone:         "UTC",
		RetrievedAt:      time.Now().UTC(),
	}, nil
}

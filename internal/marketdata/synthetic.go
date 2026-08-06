package marketdata

import (
	"context"

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

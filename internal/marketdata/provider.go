package marketdata

import (
	"context"

	"github.com/kevincouton/standard-tools-go/internal/core"
)

type Provider interface {
	Name() string
	Fetch(ctx context.Context, ticker core.Ticker, interval core.BarInterval, rng core.DateRange) ([]core.OHLCV, error)
	GetTickerInfo(ctx context.Context, ticker core.Ticker) (core.TickerInfo, error)
	GetFinancialRatios(ctx context.Context, ticker core.Ticker) (core.FinancialRatios, error)
	GetMetadata(ctx context.Context) (core.DataSetMetadata, error)
}

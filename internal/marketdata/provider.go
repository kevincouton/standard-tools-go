package marketdata

import (
	"context"

	"github.com/kevincouton/standard-tools-go/internal/core"
)

type Provider interface {
	Name() string
	Fetch(ctx context.Context, ticker core.Ticker, interval core.BarInterval, rng core.DateRange) ([]core.OHLCV, error)
}

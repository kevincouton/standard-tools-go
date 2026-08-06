package marketdata

import (
	"context"
	"fmt"

	"github.com/kevincouton/standard-tools-go/internal/core"
)

type YahooProvider struct{}

func (y *YahooProvider) Name() string { return "yahoo" }

func (y *YahooProvider) Fetch(ctx context.Context, ticker core.Ticker, interval core.BarInterval, rng core.DateRange) ([]core.OHLCV, error) {
	return nil, fmt.Errorf("%w: yahoo fetch not yet implemented", core.ErrProviderNotAvailable)
}

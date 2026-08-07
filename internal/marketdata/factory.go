package marketdata

import (
	"fmt"
	"os"

	"github.com/kevincouton/standard-tools-go/internal/core"
)

func NewProvider(name string) (Provider, error) {
	switch name {
	case "synthetic":
		return &SyntheticProvider{}, nil
	case "yahoo":
		return NewYahooProvider(), nil
	case "polygon":
		return NewPolygonProvider(os.Getenv("SQT_POLYGON_API_KEY"))
	default:
		return nil, fmt.Errorf("%w: unknown provider %s", core.ErrProviderNotAvailable, name)
	}
}

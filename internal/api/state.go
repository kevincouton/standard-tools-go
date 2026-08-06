package api

import (
	"github.com/kevincouton/standard-tools-go/internal/agent"
	"github.com/kevincouton/standard-tools-go/internal/marketdata"
)

type AppState struct {
	Dispatcher *agent.Dispatcher
	MarketData *marketdata.Service
}

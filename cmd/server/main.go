package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/kevincouton/standard-tools-go/internal/agent"
	"github.com/kevincouton/standard-tools-go/internal/api"
	"github.com/kevincouton/standard-tools-go/internal/marketdata"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cache := marketdata.NewInMemoryCache()
	svc := marketdata.NewService("synthetic", cache)
	svc.Register(&marketdata.SyntheticProvider{})
	svc.Register(&marketdata.YahooProvider{})

	state := &api.AppState{
		Dispatcher: agent.NewDispatcher(svc),
		MarketData: svc,
	}

	slog.Info("starting server", "http", 8080, "grpc", 50051)
	if err := api.Serve(context.Background(), state, 8080, 50051); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

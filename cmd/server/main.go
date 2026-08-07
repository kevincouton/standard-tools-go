package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/kevincouton/standard-tools-go/internal/agent"
	"github.com/kevincouton/standard-tools-go/internal/api"
	"github.com/kevincouton/standard-tools-go/internal/marketdata"
)

const (
	httpPort = 8080
	grpcPort = 50051
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cache := marketdata.NewInMemoryCache()
	svc := marketdata.NewService("synthetic", cache)
	svc.Register(&marketdata.SyntheticProvider{})
	svc.Register(marketdata.NewYahooProvider())

	state := &api.AppState{
		Dispatcher: agent.NewDispatcher(svc),
		MarketData: svc,
	}

	slog.Info("starting server", "http", httpPort, "grpc", grpcPort)
	if err := api.Serve(context.Background(), state, httpPort, grpcPort); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

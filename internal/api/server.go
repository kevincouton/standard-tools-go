package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Serve starts the HTTP and gRPC servers and blocks until ctx is cancelled or a
// server fails. Ports of 0 may be used to bind an ephemeral port.
func Serve(ctx context.Context, state *AppState, httpPort, grpcPort int) error {
	httpLis, err := net.Listen("tcp", fmt.Sprintf(":%d", httpPort))
	if err != nil {
		return err
	}
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		httpLis.Close()
		return err
	}
	return serve(ctx, state, httpLis, grpcLis)
}

// serve is the internal implementation that accepts pre-bound listeners. It is
// exported to the package tests so they can bind ephemeral ports without races.
func serve(ctx context.Context, state *AppState, httpLis net.Listener, grpcLis net.Listener) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpSrv := &http.Server{Handler: NewRouter(state)}
	grpcSrv := grpc.NewServer()
	hs := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcSrv, hs)
	hs.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := httpSrv.Serve(httpLis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})
	g.Go(func() error {
		return grpcSrv.Serve(grpcLis)
	})
	g.Go(func() error {
		<-ctx.Done()

		// Shut down HTTP server with a bounded grace period.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)

		// GracefulStop blocks until all RPCs finish; cap it and force-stop.
		done := make(chan struct{})
		go func() {
			grpcSrv.GracefulStop()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			grpcSrv.Stop()
		}

		return nil
	})
	return g.Wait()
}

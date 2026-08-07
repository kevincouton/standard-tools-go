package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/kevincouton/standard-tools-go/internal/audit"
	"github.com/kevincouton/standard-tools-go/internal/config"
	"github.com/kevincouton/standard-tools-go/internal/storage"
	"github.com/spf13/cobra"
)

// Dependencies can be overridden in tests.
var (
	loadConfig      = config.Load
	newStorage      func(ctx context.Context, cfg *config.Config) (audit.Storage, error)
	defaultNewStorage func(ctx context.Context, cfg *config.Config) (audit.Storage, error)
)

func init() {
	newStorage = func(ctx context.Context, cfg *config.Config) (audit.Storage, error) {
		if cfg.DatabaseURL == "" {
			return audit.NewMemoryStorage(), nil
		}
		pool, err := storage.NewPool(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}
		if err := storage.MigrateUp(pool); err != nil {
			pool.Close()
			return nil, err
		}
		return audit.NewPostgresStorage(pool), nil
	}
	defaultNewStorage = newStorage
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "sqt",
		Short: "Standard Quant Tools CLI",
	}
	root.AddCommand(newVerifyCmd())
	root.AddCommand(newReportCmd())
	return root
}

func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

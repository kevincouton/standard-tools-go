package cli

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kevincouton/standard-tools-go/internal/audit"
	"github.com/spf13/cobra"
)

func newReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "report <request-id>",
		Short: "Print an audit record as JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			store, err := newStorage(ctx, cfg)
			if err != nil {
				return fmt.Errorf("open storage: %w", err)
			}
			defer store.Close()

			record, err := store.GetByRequestID(ctx, args[0])
			if errors.Is(err, audit.ErrNotFound) {
				return fmt.Errorf("record not found: %s", args[0])
			}
			if err != nil {
				return fmt.Errorf("fetch record: %w", err)
			}

			b, err := json.MarshalIndent(record, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal record: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		},
	}
}

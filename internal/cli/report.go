package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "report <request-id>",
		Short: "Print an audit record as JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			store, err := newStorage(ctx, cfg)
			if err != nil {
				return err
			}
			record, err := store.GetByRequestID(ctx, args[0])
			if err != nil {
				return err
			}
			if record.RequestID == "" {
				return fmt.Errorf("record not found: %s", args[0])
			}
			b, err := json.MarshalIndent(record, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		},
	}
}

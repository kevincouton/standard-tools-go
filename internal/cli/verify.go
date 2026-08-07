package cli

import (
	"context"
	"fmt"

	"github.com/kevincouton/standard-tools-go/internal/audit"
	"github.com/spf13/cobra"
)

func newVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Verify the integrity of the audit chain",
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
			if err := audit.NewVerifier(store).VerifyChain(ctx); err != nil {
				return fmt.Errorf("audit chain verification failed: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "audit chain OK")
			return nil
		},
	}
}

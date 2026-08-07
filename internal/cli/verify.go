package cli

import (
	"fmt"

	"github.com/kevincouton/standard-tools-go/internal/audit"
	"github.com/spf13/cobra"
)

func newVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Verify the integrity of the audit chain",
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

			if err := audit.NewVerifier(store).VerifyChain(ctx); err != nil {
				return fmt.Errorf("audit chain verification failed: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "audit chain OK")
			return nil
		},
	}
}
